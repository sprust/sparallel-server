package sparallel_server

import (
	"github.com/google/uuid"
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
	return &Pool{
		waitingTasks:     make(map[string]*Task),
		processesPool:    make(map[string]*Process),
		runningTasks:     make(map[string]*ActiveWorker),
		runningProcesses: make(map[string]*ActiveWorker),
		finishedTasks:    make(map[string]map[string]*FinishedTask),
	}
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

	runningProcess, exists := p.runningProcesses[processUuid]

	if !exists {
		return
	}

	delete(p.runningProcesses, processUuid)
	delete(p.runningTasks, runningProcess.task.Uuid)
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

func (p *Pool) CancelTask(taskUuid string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

}

func (p *Pool) RegisterFinishedTasks(worker *ActiveWorker, finishedTask *FinishedTask) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

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

func (p *Pool) DetectFinishedTask(groupUuid string) *FinishedTask {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	group, exists := p.finishedTasks[groupUuid]

	if !exists {
		return &FinishedTask{
			Task:     nil,
			Response: "err: tasks group not found",
			IsError:  true,
		}
	}

	if len(group) == 0 {
		return &FinishedTask{
			Task:     nil,
			Response: "warn: no tasks in group",
			IsError:  false,
		}
	}

	var finishedTask *FinishedTask

	for _, task := range group {
		finishedTask = task

		break
	}

	p.DeleteFinishedTasks(finishedTask)

	return finishedTask
}
