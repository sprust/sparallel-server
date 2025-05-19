package sparallel_server

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"sparallel_server/pkg/foundation/errs"
	"sync"
	"time"
)

var server *Server

type Server struct {
	command                   string
	minWorkersNumber          int
	maxWorkersNumber          int
	workersNumberPercentScale int
	workersNumberScaleUp      int
	workersNumberScaleDown    int

	pool *Pool

	mutex sync.Mutex

	closing bool
}

func NewServer(
	command string,
	minWorkersNumber int,
	maxWorkersNumber int,
	workersNumberPercentScale int,
	workersNumberScaleUp int,
	workersNumberScaleDown int,
) *Server {
	if server != nil {
		panic("Sparallel server is already created")
	}

	slog.Info("Creating sparallel server for [" + command + "] command...")

	server = &Server{
		command:                   command,
		minWorkersNumber:          minWorkersNumber,
		maxWorkersNumber:          maxWorkersNumber,
		workersNumberPercentScale: workersNumberPercentScale,
		workersNumberScaleUp:      workersNumberScaleUp,
		workersNumberScaleDown:    workersNumberScaleDown,

		pool: NewPool(),
	}

	return server
}

func (s *Server) Start(ctx context.Context) {
	slog.Info("Starting sparallel server...")

	tickers := []func(s *Server){
		func(s *Server) {
			for !s.closing {
				s.readTaskResponses()
			}
		},
		func(s *Server) {
			for !s.closing {
				err := s.controlProcessesPool(ctx)

				if err != nil {
					panic(errs.Err(err))
				}
			}
		},
		func(s *Server) {
			for !s.closing {
				s.clearFinishedTasks()
			}
		},
		func(s *Server) {
			for !s.closing {
				s.startWaitingTasks()
			}
		},
	}

	for _, ticker := range tickers {
		go ticker(s)
	}

	go func() {
		for !s.closing {
			time.Sleep(1 * time.Second)

			var mem runtime.MemStats

			stats := SystemStats{
				NumGoroutine:  uint64(runtime.NumGoroutine()),
				AllocMiB:      float32(mem.Alloc / 1024 / 1024),
				TotalAllocMiB: float32(mem.TotalAlloc / 1024 / 1024),
				SysMiB:        float32(mem.Sys / 1024 / 1024),
				NumGC:         uint64(mem.NumGC),
			}

			slog.Debug(
				fmt.Sprintf(
					"go=%d\tAlloc=%v_MiB\tTotalAlloc=%v_MiB\tSys=%v_MiB\tNumGC=%v",
					stats.NumGoroutine,
					stats.AllocMiB,
					stats.TotalAllocMiB,
					stats.SysMiB,
					stats.NumGC,
				),
			)
		}
	}()
}

func (s *Server) AddTask(groupUuid string, unixTimeTimeout int, payload string) *Task {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	slog.Debug("Adding task to group [" + groupUuid + "]")

	return s.pool.AddTask(groupUuid, unixTimeTimeout, payload)
}

// CancelTask TODO
func (s *Server) CancelTask(taskUuid string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	slog.Debug("Cancelling task...")

	runningTasks, exists := s.pool.runningTasks[taskUuid]

	if exists {
		_ = runningTasks.process.Close()

		s.pool.DeleteProcess(runningTasks.process.Uuid)
	}

	task, exists := s.pool.waitingTasks[taskUuid]

	if exists {
		delete(s.pool.waitingTasks, taskUuid)

		group, exists := s.pool.finishedTasks[task.GroupUuid]

		if exists {
			finishedTask, exists := group[taskUuid]

			if exists {
				s.pool.DeleteFinishedTasks(finishedTask)
			}
		}
	}

	_, exists = s.pool.runningTasks[taskUuid]

	if exists {
		delete(s.pool.runningTasks, taskUuid)
	}
}

func (s *Server) DetectAnyFinishedTask(groupUuid string) *FinishedTask {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	finishedTask := s.pool.DetectAnyFinishedTask(groupUuid)

	if finishedTask.IsFinished {
		s.pool.DeleteFinishedTasks(finishedTask)
	}

	return finishedTask
}

func (s *Server) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	slog.Warn("Closing sparallel server...")

	s.closing = true

	for _, process := range s.pool.processesPool {
		_ = process.Close()

		s.pool.DeleteProcess(process.Uuid)
	}

	_ = s.pool.Close()

	return nil
}

func (s *Server) readTaskResponses() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	taskUuids := s.pool.GetRunningTaskKeys()

	for _, taskUuid := range taskUuids {
		worker := s.pool.GetRunningTask(taskUuid)

		if worker == nil {
			continue
		}

		var response string
		var isError bool

		if time.Now().Unix()-int64(worker.task.UnixTimeTimeout) < -int64(5*time.Second) {
			slog.Debug("Task [" + worker.task.Uuid + "] closing by timeout.")

			response = "err:timeout"
			isError = true

			_ = worker.process.Close()

			s.pool.DeleteProcess(worker.process.Uuid)
		} else {
			slog.Debug("Task [" + worker.task.Uuid + "] reading response from process.")

			processResponse := worker.process.Read()

			if processResponse == nil {
				continue
			}

			if processResponse.Error != nil {
				response = processResponse.Error.Error()
				isError = true

				_ = worker.process.Close()

				s.pool.DeleteProcess(worker.process.Uuid)
			} else {
				response = processResponse.Data
				isError = false
			}
		}

		finishedTask := &FinishedTask{
			Task:     worker.task,
			Response: response,
			IsError:  isError,
		}

		slog.Debug("Task [" + worker.task.Uuid + "] finished.")

		s.pool.RegisterFinishedTasks(worker, finishedTask)
	}
}

func (s *Server) controlProcessesPool(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	processUuids := s.pool.GetProcessPoolKeys()

	for _, processUuid := range processUuids {
		process := s.pool.GetProcessPool(processUuid)

		if process == nil {
			continue
		}

		if !process.IsRunning() {
			slog.Debug("Process[ " + processUuid + "] is not running. Removing it from pool.")

			_ = process.Close()

			s.pool.DeleteProcess(processUuid)
		}
	}

	needWorkersNumber := s.minWorkersNumber

	for len(s.pool.processesPool) < needWorkersNumber {
		newProcess, err := CreateProcess(
			ctx,
			s.command,
			func(processUuid string) {
				s.pool.DeleteProcess(processUuid)
			},
		)

		if err != nil {
			return errs.Err(err)
		}

		slog.Debug("Process [" + newProcess.Uuid + "] created.")

		s.pool.AddProcess(newProcess)
	}

	return nil
}

func (s *Server) clearFinishedTasks() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	groupUuids := s.pool.GetFinishedGroupKeys()

	for _, groupUuid := range groupUuids {
		taskUuids := s.pool.GetFinishedTaskKeys(groupUuid)

		for _, taskUuid := range taskUuids {
			finishedTask := s.pool.FindFinishedTask(groupUuid, taskUuid)

			if finishedTask == nil {
				continue
			}

			if time.Now().Unix()-int64(finishedTask.Task.UnixTimeTimeout) < -int64(20*time.Second) {
				s.pool.DeleteFinishedTasks(finishedTask)

				slog.Debug("Finished task [" + finishedTask.Task.Uuid + "] deleted.")
			}
		}
	}
}

func (s *Server) startWaitingTasks() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	activeWorkers := s.pool.CreateActiveWorkers()

	for _, activeWorker := range activeWorkers {
		err := activeWorker.process.Write(activeWorker.task.Payload)

		if err != nil {
			slog.Error("Failed to write to process: " + err.Error())

			_ = activeWorker.process.Close()

			s.pool.DeleteProcess(activeWorker.process.Uuid)
		}
	}
}
