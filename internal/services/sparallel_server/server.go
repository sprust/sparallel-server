package sparallel_server

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"log/slog"
	"runtime"
	"sparallel_server/internal/services/sparallel_server/processes"
	"sparallel_server/internal/services/sparallel_server/tasks"
	"sparallel_server/internal/services/sparallel_server/workers"
	"sparallel_server/pkg/foundation/errs"
	"sync/atomic"
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

	workers *workers.Workers
	tasks   *tasks.Tasks

	closing atomic.Bool

	tickersCtx       context.Context
	tickersCtxCancel context.CancelFunc
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

		workers: workers.NewWorkers(),
		tasks:   tasks.NewTasks(),
	}

	return server
}

func (s *Server) Start(ctx context.Context) {
	slog.Info("Starting sparallel server...")

	s.tickersCtx, s.tickersCtxCancel = context.WithCancel(ctx)

	tickers := []func(ctx context.Context, s *Server){
		func(ctx context.Context, s *Server) {
			err := s.tickControlWorkers(ctx)

			if err != nil {
				panic(errs.Err(err))
			}
		},
		func(ctx context.Context, s *Server) {
			s.tickClearFinishedTasks()
		},
		func(ctx context.Context, s *Server) {
			s.tickHandleTasks(ctx)
		},
		func(ctx context.Context, s *Server) {
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
		},
	}

	for _, ticker := range tickers {
		go func(ctx context.Context, ticker func(ctx context.Context, s *Server)) {
			for !s.closing.Load() {
				ticker(ctx, s)
			}
		}(s.tickersCtx, ticker)
	}
}

func (s *Server) AddTask(groupUuid string, unixTimeout int, payload string) *tasks.Task {
	slog.Debug("Adding task to group [" + groupUuid + "]")

	newTask := &tasks.Task{
		GroupUuid:   groupUuid,
		Uuid:        uuid.New().String(),
		UnixTimeout: unixTimeout,
		Payload:     payload,
	}

	s.tasks.AddWaiting(newTask)

	return newTask
}

func (s *Server) DetectAnyFinishedTask(groupUuid string) *tasks.Task {
	finishedTask := s.tasks.TakeFinished(groupUuid)

	if finishedTask == nil {
		return &tasks.Task{
			IsFinished: false,
		}
	}

	return finishedTask
}

func (s *Server) Close() error {
	slog.Warn("Closing sparallel server...")

	s.closing.Store(true)

	s.tickersCtxCancel()

	_ = s.workers.Close()

	return nil
}

func (s *Server) tickControlWorkers(ctx context.Context) error {
	needWorkersNumber := s.minWorkersNumber

	for s.workers.Count() < needWorkersNumber {
		newProcess, err := processes.CreateProcess(
			ctx,
			s.command,
			func(processUuid string) {
				task := s.workers.DeleteAndGetTask(processUuid)

				if task != nil {
					s.tasks.AddWaiting(task)
				}
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

func (s *Server) tickClearFinishedTasks() {
	s.tasks.FlushRottenFinished()
}

func (s *Server) tickHandleTasks(ctx context.Context) {
	task := s.tasks.TakeWaiting()

	if task == nil {
		return
	}

	go func(_ context.Context, task *tasks.Task) {
		s.handleTask(task)
	}(ctx, task)
}

func (s *Server) handleTask(task *tasks.Task) {
	worker := s.workers.Take()

	if worker == nil {
		s.tasks.AddWaiting(task)

		return
	}

	s.workers.Busy(worker, task)

	process := worker.GetProcess()

	err := process.Write(task.Payload)

	if err != nil {
		s.tasks.AddWaiting(task)

		s.workers.DeleteAndGetTask(process.Uuid)

		_ = process.Close()

		return
	}

	for {
		response := process.Read()

		if response == nil {
			if task.IsTimeout() {
				task.IsFinished = true
				task.Response = "timeout"
				task.IsError = true

				s.tasks.AddFinished(task)

				break
			}

			continue
		}

		if response.Error != nil {
			s.workers.DeleteAndGetTask(process.Uuid)

			_ = process.Close()

			s.tasks.AddWaiting(task)

			task.IsFinished = true
			task.Response = response.Error.Error()
			task.IsError = true

			s.tasks.AddFinished(task)

			break
		}

		task.IsFinished = true
		task.Response = response.Data
		task.IsError = false

		s.tasks.AddFinished(task)

		break
	}

	s.workers.Free(worker)
}
