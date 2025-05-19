package sparallel_server

import (
	"context"
	"fmt"
	"log/slog"
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
			for {
				s.readTaskResponses()
			}
		},
		func(s *Server) {
			for {
				err := s.controlProcessesPool(ctx)

				if err != nil {
					panic(errs.Err(err))
				}
			}
		},
		func(s *Server) {
			for {
				s.clearFinishedTasks()
			}
		},
		func(s *Server) {
			for {
				s.startWaitingTasks()
			}
		},
	}

	for _, ticker := range tickers {
		go ticker(s)
	}
}

func (s *Server) AddTask(groupUuid string, unixTimeTimeout int, payload string) *Task {
	slog.Info(
		fmt.Sprintf(
			"Adding task to group [%s]",
			groupUuid,
		))

	return s.pool.AddTask(groupUuid, unixTimeTimeout, payload)
}

// CancelTask TODO
func (s *Server) CancelTask(taskUuid string) {
	slog.Info("Cancelling task...")

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
	slog.Info("Detecting finished task for group [" + groupUuid + "]")

	finishedTask := s.pool.DetectAnyFinishedTask(groupUuid)

	if finishedTask.IsFinished {
		s.pool.DeleteFinishedTasks(finishedTask)
	}

	return finishedTask
}

func (s *Server) Close() error {
	slog.Warn("Closing sparallel server...")

	for _, process := range s.pool.processesPool {
		_ = process.Close()
	}

	_ = s.pool.Close()

	return nil
}

func (s *Server) readTaskResponses() {
	taskUuids := s.pool.GetRunningTaskKeys()

	for _, taskUuid := range taskUuids {
		worker := s.pool.GetRunningTask(taskUuid)

		if worker == nil {
			continue
		}

		var response string
		var isError bool

		if time.Now().Unix()-int64(worker.task.UnixTimeTimeout) < -int64(5*time.Second) {
			response = "err:timeout"
			isError = true

			_ = worker.process.Close()
		} else {
			slog.Warn("Task [" + worker.task.Uuid + "] reading response from process.")

			processResponse := worker.process.Read()

			if processResponse == nil {
				continue
			}

			if processResponse.Error != nil {
				response = processResponse.Error.Error()
				isError = true
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

		slog.Warn("Task [" + worker.task.Uuid + "] finished.")

		s.pool.RegisterFinishedTasks(worker, finishedTask)
	}
}

func (s *Server) controlProcessesPool(ctx context.Context) error {
	processUuids := s.pool.GetProcessPoolKeys()

	for _, processUuid := range processUuids {
		process := s.pool.GetProcessPool(processUuid)

		if process == nil {
			continue
		}

		if !process.IsRunning() {
			slog.Warn("Process[ " + processUuid + "] is not running. Removing it from pool.")

			s.pool.DeleteProcess(processUuid)
		}
	}

	needWorkersNumber := s.minWorkersNumber

	for len(s.pool.processesPool) < needWorkersNumber {
		newProcess, err := CreateProcess(ctx, s.command)

		if err != nil {
			return errs.Err(err)
		}

		slog.Info("Process [" + newProcess.Uuid + "] created.")

		s.pool.AddProcess(newProcess)
	}

	return nil
}

func (s *Server) clearFinishedTasks() {
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

				slog.Warn("Finished task [" + finishedTask.Task.Uuid + "] deleted.")
			}
		}
	}
}

func (s *Server) startWaitingTasks() {
	activeWorkers := s.pool.CreateActiveWorkers()

	for _, activeWorker := range activeWorkers {
		err := activeWorker.process.Write(activeWorker.task.Payload)

		if err != nil {
			panic(err)
		}
	}
}
