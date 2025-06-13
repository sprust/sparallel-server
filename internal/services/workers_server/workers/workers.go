package workers

import (
	"github.com/google/uuid"
	"log/slog"
	"sparallel_server/internal/services/workers_server/processes"
	"sparallel_server/internal/services/workers_server/tasks"
	"strconv"
	"time"
)

func NewWorkers() *Workers {
	return &Workers{
		pw:   make(map[string]*Worker),
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

	w.pw[process.Uuid] = newWorker

	w.free[workerUuid] = newWorker

	w.totalCount.Add(1)
	w.freeCount.Add(1)

	w.incCount(&w.addedCount)
}

func (w *Workers) Take(task *tasks.Task) *Worker {
	if w.freeCount.Load() == 0 || w.closing.Load() {
		return nil
	}

	w.mutex.Lock()
	defer w.mutex.Unlock()

	var selectedWorker *Worker

	for _, worker := range w.free {
		delete(w.free, worker.uuid)

		w.freeCount.Add(-1)

		selectedWorker = worker

		break
	}

	if selectedWorker == nil {
		return nil
	}

	selectedWorker.task = task

	w.busy[selectedWorker.uuid] = selectedWorker

	w.busyCount.Add(1)

	w.incCount(&w.tookCount)

	return selectedWorker
}

func (w *Workers) Free(worker *Worker) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	worker.task = nil

	if worker.reload {
		w.deleteByProcessUuid(worker.process.Uuid)

		_ = worker.process.Close()

		return
	}

	delete(w.busy, worker.uuid)

	w.busyCount.Add(-1)

	w.free[worker.uuid] = worker

	w.freeCount.Add(1)

	w.incCount(&w.freedCount)
}

func (w *Workers) DeleteByProcess(processUuid string) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.deleteByProcessUuid(processUuid)
}

func (w *Workers) DeleteByGroup(groupUuid string) []*processes.Process {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	var deletedProcesses []*processes.Process

	for _, worker := range w.pw {
		if worker.task == nil {
			continue
		}

		if worker.task.GroupUuid == groupUuid {
			deletedProcess := w.deleteByProcessUuid(worker.process.Uuid)

			deletedProcesses = append(deletedProcesses, deletedProcess)
		}
	}

	return deletedProcesses
}

func (w *Workers) KillAnyFree() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	for _, worker := range w.free {
		w.deleteByProcessUuid(worker.process.Uuid)

		_ = worker.process.Close()

		break
	}
}

func (w *Workers) Reload() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	for _, worker := range w.busy {
		worker.reload = true
	}

	for _, worker := range w.free {
		w.deleteByProcessUuid(worker.process.Uuid)

		_ = worker.process.Close()
	}
}

func (w *Workers) HasProcess(pid int) bool {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	for _, worker := range w.pw {
		if worker.process.Cmd.Process.Pid == pid {
			return true
		}
	}

	return false
}

func (w *Workers) Close() error {
	slog.Warn("Closing workers...")

	w.closing.Store(true)

	triesCount := 5

	for w.GetBusyCount() > 0 && triesCount > 0 {
		slog.Warn("Waiting for workers to close [" + strconv.Itoa(triesCount) + "]...")

		time.Sleep(1 * time.Second)

		triesCount -= 1
	}

	w.mutex.Lock()
	defer w.mutex.Unlock()

	for _, worker := range w.pw {
		_ = worker.process.Close()
	}

	w.pw = make(map[string]*Worker)
	w.free = make(map[string]*Worker)
	w.busy = make(map[string]*Worker)

	return nil
}

func (w *Workers) deleteByProcessUuid(processUuid string) *processes.Process {
	worker, exists := w.pw[processUuid]

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

	delete(w.pw, processUuid)

	w.totalCount.Add(-1)

	w.incCount(&w.deletedCount)

	return worker.process
}
