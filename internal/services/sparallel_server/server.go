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

	countAddTask                atomic.Uint64
	countDetectAnyFinishedTask  atomic.Uint64
	countTickControlWorkers     atomic.Uint64
	countTickClearFinishedTasks atomic.Uint64
	countTickHandleTasks        atomic.Uint64
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
					"work:\ttot=%d\twf=%v_MiB\twf=%v",
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

			slog.Debug(
				fmt.Sprintf(
					"calls:\tat=%d\tdetFin=%v\ttCW=%v\ttCFT=%v\ttHT=%v",
					s.countAddTask.Load(),
					s.countDetectAnyFinishedTask.Load(),
					s.countTickControlWorkers.Load(),
					s.countTickClearFinishedTasks.Load(),
					s.countTickHandleTasks.Load(),
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

	go s.countAddTask.Add(1)

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
	go s.countDetectAnyFinishedTask.Add(1)

	finishedTask := s.tasks.TakeFinished(groupUuid)

	if finishedTask == nil {
		return &tasks.Task{
			IsFinished: false,
		}
	}

	return finishedTask
}

func (s *Server) Close() error {
	s.closing.Store(true)

	s.tickersCtxCancel()

	slog.Warn("Closing sparallel server...")

	_ = s.workers.Close()

	return nil
}

func (s *Server) tickControlWorkers(ctx context.Context) error {
	go s.countTickControlWorkers.Add(1)

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
	go s.countTickClearFinishedTasks.Add(1)

	s.tasks.FlushRottenTasks()
}

func (s *Server) tickHandleTasks(ctx context.Context) {
	go s.countTickHandleTasks.Add(1)

	task := s.tasks.TakeWaiting()

	if task == nil {
		return
	}

	go func(_ context.Context, task *tasks.Task) {
		s.handleTask(task)
	}(ctx, task)
}

func (s *Server) handleTask(task *tasks.Task) {
	slog.Debug("Handling task [" + task.Uuid + "]")

	worker := s.workers.Take(task)

	if worker == nil {
		slog.Debug("Not found worker for task [" + task.Uuid + "]")

		s.tasks.AddWaiting(task)

		return
	}

	process := worker.GetProcess()

	err := process.Write(task.Payload)

	if err != nil {
		slog.Debug("Error start task [" + task.Uuid + "]")

		s.tasks.AddWaiting(task)

		s.workers.DeleteAndGetTask(process.Uuid)

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
			s.workers.DeleteAndGetTask(process.Uuid)

			_ = process.Close()

			task.IsFinished = true
			task.Response = response.Error.Error()
			task.IsError = true

			s.tasks.AddFinished(task)

			return
		}

		task.IsFinished = true
		task.Response = response.Data
		task.IsError = false

		s.workers.Free(worker)
		s.tasks.AddFinished(task)

		break
	}
}
