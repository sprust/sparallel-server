package tasks

import (
	"log/slog"
	"strconv"
)

func NewTasks() *Tasks {
	return &Tasks{
		waiting:  NewSubTasks(),
		finished: NewSubTasks(),
	}
}

func (t *Tasks) AddWaiting(task *Task) {
	slog.Debug("Task [" + task.TaskUuid + "] waiting")

	t.waiting.AddTask(task)
}

func (t *Tasks) TakeWaiting() *Task {
	task := t.waiting.Pop()

	if task == nil {
		return nil
	}

	slog.Debug("Task [" + task.TaskUuid + "] taken")

	return task
}

func (t *Tasks) AddFinished(task *Task) {
	slog.Debug("Task [" + task.TaskUuid + "] finished")

	t.finished.AddTask(task)
}

func (t *Tasks) TakeFinished(groupUuid string) *Task {
	return t.finished.TakeFirstByGroupUuid(groupUuid)
}

func (t *Tasks) FlushRottenTasks() {
	var deletedCount int

	deletedCount = t.waiting.FlushFirstRotten()

	if deletedCount > 0 {
		slog.Debug("Flushed rotten waiting tasks: " + strconv.Itoa(deletedCount))
	}

	deletedCount = t.finished.FlushFirstRotten()

	if deletedCount > 0 {
		slog.Debug("Flushed rotten finished tasks: " + strconv.Itoa(deletedCount))
	}
}

func (t *Tasks) DeleteGroup(groupUuid string) {
	t.waiting.DeleteGroup(groupUuid)
	t.finished.DeleteGroup(groupUuid)
}

func (t *Tasks) DeleteTask(task *Task) {
	t.waiting.DeleteTask(task)
	t.finished.DeleteTask(task)
}

func (t *Tasks) WaitingCount() int {
	return t.waiting.Count()
}

func (t *Tasks) FinishedCount() int {
	return t.finished.Count()
}
