package workers

import (
	"github.com/google/uuid"
	"sparallel_server/internal/services/sparallel_server/processes"
	"sparallel_server/internal/services/sparallel_server/tasks"
)

func NewWorkers() *Workers {
	return &Workers{
		all:  make(map[string]*Worker),
		free: make(map[string]*Worker),
		busy: make(map[string]*Worker),
	}
}

func (w *Workers) Add(process *processes.Process) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	workerUuid := uuid.New().String()

	newWorker := &Worker{
		uuid:    workerUuid,
		process: process,
	}

	w.all[workerUuid] = newWorker
	w.free[workerUuid] = newWorker

	w.totalCount.Add(1)
	w.freeCount.Add(1)
}

func (w *Workers) Take() *Worker {
	if w.freeCount.Load() == 0 {
		return nil
	}

	w.mutex.Lock()
	defer w.mutex.Unlock()

	for wu, worker := range w.free {
		delete(w.free, wu)

		w.freeCount.Add(-1)

		return worker
	}

	return nil
}

func (w *Workers) Busy(worker *Worker, task *tasks.Task) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	worker.task = task

	w.busy[worker.uuid] = worker

	w.busyCount.Add(1)
}

func (w *Workers) Free(worker *Worker) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	worker.task = nil

	delete(w.busy, worker.uuid)

	w.busyCount.Add(-1)

	w.free[worker.uuid] = worker

	w.freeCount.Add(1)
}

func (w *Workers) DeleteAndGetTask(processUuid string) *tasks.Task {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	worker, exists := w.all[processUuid]

	if !exists {
		return nil
	}

	if _, exists = w.free[worker.uuid]; exists {
		delete(w.free, worker.uuid)

		w.freeCount.Add(-1)
	}

	if _, exists = w.busy[worker.uuid]; exists {
		delete(w.busy, worker.uuid)

		w.busyCount.Add(-1)
	}

	delete(w.all, worker.uuid)

	w.totalCount.Add(-1)

	return worker.task
}

func (w *Workers) Count() int {
	return int(w.totalCount.Load())
}

func (w *Workers) BusyCount() int {
	return int(w.busyCount.Load())
}

func (w *Workers) FreeCount() int {
	return int(w.freeCount.Load())
}

func (w *Workers) Close() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	for _, worker := range w.all {
		_ = worker.process.Close()
	}

	w.all = make(map[string]*Worker)
	w.free = make(map[string]*Worker)
	w.busy = make(map[string]*Worker)

	return nil
}
