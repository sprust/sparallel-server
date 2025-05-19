package sparallel_server

import (
	"github.com/google/uuid"
	"log/slog"
	"sparallel_server/internal/helpers"
	"sync"
)

type Pool struct {
	mutex sync.Mutex

	// map[taskUuid]
	waitingTasks map[string]*Task
	// map[processUuid]
	processesPool map[string]*Process

	// map[taskUuid]
	runningTasks map[string]*ActiveWorker
	// map[processUuid]
	runningProcesses map[string]*ActiveWorker

	// map[groupUuid]map[taskUuid]
	finishedTasks map[string]map[string]*FinishedTask
}

func NewPool() *Pool {
	pool := &Pool{}

	pool.flush()

	return pool
}

func (p *Pool) GetRunningTaskKeys() []string {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return helpers.GetMapKeys(p.runningTasks)
}

func (p *Pool) GetRunningTask(taskUuid string) *ActiveWorker {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.runningTasks[taskUuid]
}

func (p *Pool) GetProcessPoolKeys() []string {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return helpers.GetMapKeys(p.processesPool)
}

func (p *Pool) GetProcessPool(processUuid string) *Process {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.processesPool[processUuid]
}

func (p *Pool) GetFinishedGroupKeys() []string {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return helpers.GetMapKeys(p.finishedTasks)
}

func (p *Pool) GetFinishedTaskKeys(groupUuid string) []string {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return helpers.GetMapKeys(p.finishedTasks[groupUuid])
}

func (p *Pool) FindFinishedTask(groupUuid string, taskUuid string) *FinishedTask {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.finishedTasks[groupUuid][taskUuid]
}

func (p *Pool) AddProcess(process *Process) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.processesPool[process.Uuid] = process
}

func (p *Pool) DeleteProcess(processUuid string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	delete(p.processesPool, processUuid)

	_, exists := p.runningProcesses[processUuid]

	if !exists {
		return
	}

	delete(p.runningProcesses, processUuid)
}

func (p *Pool) AddTask(groupUuid string, unixTimeTimeout int, payload string) *Task {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	newTask := &Task{
		GroupUuid:       groupUuid,
		Uuid:            uuid.New().String(),
		UnixTimeTimeout: unixTimeTimeout,
		Payload:         payload,
	}

	p.waitingTasks[newTask.Uuid] = newTask

	return newTask
}

func (p *Pool) CancelTask() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
}

func (p *Pool) RegisterFinishedTasks(worker *ActiveWorker, finishedTask *FinishedTask) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	_, exists := p.finishedTasks[worker.task.GroupUuid]

	if !exists {
		p.finishedTasks[worker.task.GroupUuid] = make(map[string]*FinishedTask)
	}

	p.finishedTasks[worker.task.GroupUuid][finishedTask.Task.Uuid] = finishedTask

	delete(p.runningTasks, worker.task.Uuid)
	delete(p.runningProcesses, worker.process.Uuid)
}

func (p *Pool) DeleteFinishedTasks(finishedTask *FinishedTask) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	groupUuid := finishedTask.Task.GroupUuid

	group, exists := p.finishedTasks[groupUuid]

	if !exists {
		return
	}

	delete(group, finishedTask.Task.Uuid)

	if len(p.finishedTasks[groupUuid]) == 0 {
		delete(p.finishedTasks, groupUuid)
	}
}

func (p *Pool) CreateActiveWorkers() []*ActiveWorker {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	var activeWorkers []*ActiveWorker

	taskUuids := helpers.GetMapKeys(p.waitingTasks)

	if len(taskUuids) == 0 {
		return activeWorkers
	}

	var freeProcesses []*Process

	for _, processUuid := range p.processesPool {
		_, exists := p.runningProcesses[processUuid.Uuid]

		if exists {
			continue
		}

		freeProcesses = append(freeProcesses, processUuid)
	}

	if len(freeProcesses) == 0 {
		return activeWorkers
	}

	if len(taskUuids) > len(freeProcesses) {
		taskUuids = taskUuids[0:len(freeProcesses)]
	}

	for index, taskUuid := range taskUuids {
		task := p.waitingTasks[taskUuid]
		process := freeProcesses[index]

		activeWorker := &ActiveWorker{
			task:    task,
			process: process,
		}

		activeWorkers = append(activeWorkers, activeWorker)

		delete(p.waitingTasks, taskUuid)

		p.runningTasks[task.Uuid] = activeWorker
		p.runningProcesses[process.Uuid] = activeWorker
	}

	return activeWorkers
}

func (p *Pool) DetectAnyFinishedTask(groupUuid string) *FinishedTask {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	group, exists := p.finishedTasks[groupUuid]

	if !exists {
		return &FinishedTask{
			Task:       nil,
			IsFinished: false,
			Response:   "task group not finished",
			IsError:    false,
		}
	}

	var finishedTask *FinishedTask

	for _, task := range group {
		finishedTask = task

		break
	}

	if finishedTask == nil {
		return &FinishedTask{
			Task:       nil,
			IsFinished: false,
			Response:   "task not finished",
			IsError:    false,
		}
	}

	finishedTask.IsFinished = true

	return finishedTask
}

func (p *Pool) flush() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.waitingTasks = make(map[string]*Task)
	p.processesPool = make(map[string]*Process)
	p.runningTasks = make(map[string]*ActiveWorker)
	p.runningProcesses = make(map[string]*ActiveWorker)
	p.finishedTasks = make(map[string]map[string]*FinishedTask)
}

func (p *Pool) killAllProcesses() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	slog.Warn("Killing all processes...")

	for _, process := range p.processesPool {
		_ = process.Close()
	}
}

func (p *Pool) Close() error {
	p.killAllProcesses()

	p.flush()

	return nil
}
