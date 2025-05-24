package workers_server

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"runtime"
	"sparallel_server/internal/services/workers_server/processes"
	"sparallel_server/internal/services/workers_server/tasks"
	"sparallel_server/internal/services/workers_server/workers"
	"sparallel_server/pkg/foundation/errs"
	"sync/atomic"
	"time"
)

var service *Service

type Service struct {
	command                   string
	minWorkersNumber          int
	maxWorkersNumber          int
	workersNumberPercentScale int
	workersNumberScaleUp      int
	workersNumberScaleDown    int

	workers *workers.Workers
	tasks   *tasks.Tasks

	closing atomic.Bool

	tickersCtx       context.Context
	tickersCtxCancel context.CancelFunc
}

func NewService(
	command string,
	minWorkersNumber int,
	maxWorkersNumber int,
	workersNumberPercentScale int,
	workersNumberScaleUp int,
	workersNumberScaleDown int,
) *Service {
	if service != nil {
		panic("service is already created")
	}

	slog.Info("Creating workers service for [" + command + "] command...")

	service = &Service{
		command:                   command,
		minWorkersNumber:          minWorkersNumber,
		maxWorkersNumber:          maxWorkersNumber,
		workersNumberPercentScale: workersNumberPercentScale,
		workersNumberScaleUp:      workersNumberScaleUp,
		workersNumberScaleDown:    workersNumberScaleDown,

		workers: workers.NewWorkers(),
		tasks:   tasks.NewTasks(),

		closing: atomic.Bool{},
	}

	return service
}

func (s *Service) Start(ctx context.Context) {
	slog.Info("Starting workers service...")

	s.tickersCtx, s.tickersCtxCancel = context.WithCancel(ctx)

	tickers := []func(ctx context.Context, s *Service){
		func(ctx context.Context, s *Service) {
			err := s.tickControlWorkers(ctx)

			if err != nil {
				panic(errs.Err(err))
			}

			time.Sleep(1 * time.Second)
		},
		func(ctx context.Context, s *Service) {
			s.tickClearFinishedTasks()

			time.Sleep(5 * time.Second)
		},
		func(ctx context.Context, s *Service) {
			s.tickHandleTasks(ctx)
		},
		func(ctx context.Context, s *Service) {
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
					"sys:\tgo=%d\tAlloc=%v_MiB\tTotAlloc=%v_MiB\tSys=%v_MiB\tNumGC=%v",
					stats.NumGoroutine,
					stats.AllocMiB,
					stats.TotalAllocMiB,
					stats.SysMiB,
					stats.NumGC,
				),
			)

			slog.Debug(
				fmt.Sprintf(
					"work:\ttot=%d\tfree=%v\tbusy=%v",
					s.workers.Count(),
					s.workers.FreeCount(),
					s.workers.BusyCount(),
				),
			)

			slog.Debug(
				fmt.Sprintf(
					"tasks:\twait=%d\tfin=%v",
					s.tasks.WaitingCount(),
					s.tasks.FinishedCount(),
				),
			)
		},
	}

	for _, ticker := range tickers {
		go func(ctx context.Context, ticker func(ctx context.Context, s *Service)) {
			for !s.closing.Load() {
				ticker(ctx, s)
			}
		}(s.tickersCtx, ticker)
	}
}

func (s *Service) AddTask(groupUuid string, taskUuid string, unixTimeout int, payload string) *tasks.Task {
	slog.Debug("Adding task [" + taskUuid + "] to group [" + groupUuid + "]")

	newTask := &tasks.Task{
		GroupUuid:   groupUuid,
		TaskUuid:    taskUuid,
		UnixTimeout: unixTimeout,
		Payload:     payload,
	}

	s.tasks.AddWaiting(newTask)

	return newTask
}

func (s *Service) DetectAnyFinishedTask(groupUuid string) *tasks.Task {
	finishedTask := s.tasks.TakeFinished(groupUuid)

	if finishedTask == nil {
		return &tasks.Task{
			GroupUuid:  groupUuid,
			IsFinished: false,
		}
	}

	return finishedTask
}

// CancelGroup TODO: test
func (s *Service) CancelGroup(groupUuid string) {
	s.tasks.DeleteGroup(groupUuid)

	deletedProcesses := s.workers.DeleteByGroup(groupUuid)

	for _, deletedProcess := range deletedProcesses {
		if deletedProcess != nil {
			_ = deletedProcess.Close()
		}
	}
}

func (s *Service) Close() error {
	s.closing.Store(true)

	s.tickersCtxCancel()

	slog.Warn("Closing workers service...")

	_ = s.workers.Close()

	return nil
}

func (s *Service) tickControlWorkers(ctx context.Context) error {
	needWorkersNumber := s.minWorkersNumber

	for s.workers.Count() < needWorkersNumber {
		newProcess, err := processes.CreateProcess(
			ctx,
			s.command,
			func(processUuid string, cmd *exec.Cmd) {
				slog.Warn("Process [" + processUuid + "] finished: " + cmd.ProcessState.String())

				s.workers.DeleteByProcess(processUuid)
			},
		)

		if err != nil {
			return errs.Err(err)
		}

		slog.Debug("Process [" + newProcess.Uuid + "] created.")

		s.workers.Add(newProcess)
	}

	return nil
}

func (s *Service) tickClearFinishedTasks() {
	s.tasks.FlushRottenTasks()
}

func (s *Service) tickHandleTasks(ctx context.Context) {
	if s.workers.FreeCount() == 0 {
		return
	}

	task := s.tasks.TakeWaiting()

	if task == nil {
		return
	}

	go func(_ context.Context, task *tasks.Task) {
		s.handleTask(task)
	}(ctx, task)
}

func (s *Service) handleTask(task *tasks.Task) {
	slog.Debug("Handling task [" + task.TaskUuid + "]")

	worker := s.workers.Take(task)

	if worker == nil {
		slog.Debug("Not found worker for task [" + task.TaskUuid + "]")

		s.tasks.AddWaiting(task)

		return
	}

	process := worker.GetProcess()

	err := process.Write(task.Payload)

	if err != nil {
		slog.Debug("Error start task [" + task.TaskUuid + "]. Re waiting.")

		s.tasks.AddWaiting(task)

		s.workers.DeleteByProcess(process.Uuid)

		_ = process.Close()

		return
	}

	for {
		if task.IsTimeout() {
			task.IsFinished = true
			task.Response = "timeout"
			task.IsError = true

			s.tasks.AddFinished(task)

			s.workers.Free(worker)

			break
		}

		response := process.Read()

		if response == nil {
			continue
		}

		if response.Error != nil {
			s.workers.DeleteByProcess(process.Uuid)

			_ = process.Close()

			task.IsFinished = true
			task.Response = response.Error.Error()
			task.IsError = true

			s.tasks.AddFinished(task)

			break
		}

		task.IsFinished = true
		task.Response = response.Data
		task.IsError = false

		s.workers.Free(worker)
		s.tasks.AddFinished(task)

		break
	}
}
