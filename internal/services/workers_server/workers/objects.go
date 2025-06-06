package workers

import (
	"sparallel_server/internal/services/workers_server/processes"
	"sparallel_server/internal/services/workers_server/tasks"
	"sync"
	"sync/atomic"
)

type Workers struct {
	mutex sync.Mutex

	// map[ProcessUuid]
	pw map[string]*Worker

	// map[WorkerUuid]
	free map[string]*Worker
	// map[WorkerUuid]
	busy map[string]*Worker

	totalCount atomic.Int32
	busyCount  atomic.Int32
	freeCount  atomic.Int32

	closing atomic.Bool
}

type Worker struct {
	uuid    string
	process *processes.Process
	task    *tasks.Task
	reload  bool
}

func (w *Worker) GetProcess() *processes.Process {
	return w.process
}
