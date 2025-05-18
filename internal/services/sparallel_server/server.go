package sparallel_server

import (
	"context"
	"fmt"
	"log/slog"
	"sparallel_server/internal/helpers"
	"sparallel_server/pkg/foundation/errs"
	"sync"
	"time"
)

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
	slog.Info("Creating sparallel server for [" + command + "] command...")

	return &Server{
		command:                   command,
		minWorkersNumber:          minWorkersNumber,
		maxWorkersNumber:          maxWorkersNumber,
		workersNumberPercentScale: workersNumberPercentScale,
		workersNumberScaleUp:      workersNumberScaleUp,
		workersNumberScaleDown:    workersNumberScaleDown,

		pool: NewPool(),
	}
}

func (s *Server) Start(ctx context.Context) {
	slog.Info("Starting sparallel server...")

	go func(s *Server) {
		for {
			err := s.tick(ctx)

			if err != nil {
				panic(err)
			}
		}
	}(s)
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

func (s *Server) tick(ctx context.Context) error {
	s.readTaskResponses()

	err := s.controlProcessesPool(ctx)

	if err != nil {
		return errs.Err(err)
	}

	s.clearFinishedTasks()

	s.startWaitingTasks()

	return nil
}

func (s *Server) readTaskResponses() {
	taskUuids := helpers.GetMapKeys(s.pool.runningTasks)

	for _, taskUuid := range taskUuids {
		worker := s.pool.runningTasks[taskUuid]

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
	processUuids := helpers.GetMapKeys(s.pool.processesPool)

	for _, processUuid := range processUuids {
		process := s.pool.processesPool[processUuid]

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
	groupUuids := helpers.GetMapKeys(s.pool.finishedTasks)

	for _, groupUuid := range groupUuids {
		taskUuids := helpers.GetMapKeys(s.pool.finishedTasks[groupUuid])

		for _, taskUuid := range taskUuids {
			finishedTask := s.pool.finishedTasks[groupUuid][taskUuid]

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
