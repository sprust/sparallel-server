package tasks

import (
	"log/slog"
	"sparallel_server/pkg/foundation/helpers"
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

	helpers.IncInt64Async(&t.addedTotalCount)
}

func (t *Tasks) ReAddWaiting(task *Task) {
	slog.Debug("Task [" + task.TaskUuid + "] waiting again")

	t.waiting.AddTask(task)

	helpers.IncInt64Async(&t.reAddedTotalCount)
}

func (t *Tasks) TakeWaiting() *Task {
	task := t.waiting.Pop()

	if task == nil {
		return nil
	}

	helpers.IncInt64Async(&t.tookTotalCount)

	slog.Debug("Task [" + task.TaskUuid + "] taken")

	return task
}

func (t *Tasks) AddFinished(task *Task) {
	slog.Debug("Task [" + task.TaskUuid + "] finished")

	t.finished.AddTask(task)

	helpers.IncInt64Async(&t.finishedTotalCount)

	if task.IsError {
		helpers.IncInt64Async(&t.errorTotalCount)
	} else {
		helpers.IncInt64Async(&t.successTotalCount)
	}
}

func (t *Tasks) TakeFinished(groupUuid string) *Task {
	return t.finished.TakeFirstByGroupUuid(groupUuid)
}

func (t *Tasks) FlushRottenTasks() {
	var deletedCount int

	deletedCount = t.waiting.FlushFirstRotten()

	if deletedCount > 0 {
		slog.Debug("Flushed rotten waiting tasks: " + strconv.Itoa(deletedCount))

		helpers.IncInt64AsyncDelta(&t.timeoutTotalCount, deletedCount)
	}

	deletedCount = t.finished.FlushFirstRotten()

	if deletedCount > 0 {
		slog.Debug("Flushed rotten finished tasks: " + strconv.Itoa(deletedCount))

		helpers.IncInt64AsyncDelta(&t.timeoutTotalCount, deletedCount)
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
