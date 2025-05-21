package tasks

func NewTasks() *Tasks {
	return &Tasks{
		waiting:  NewSubTasks(),
		finished: NewSubTasks(),
	}
}

func (t *Tasks) AddWaiting(task *Task) {
	addTaskToGroup(task, t.waiting)
}

func (t *Tasks) TakeWaiting() *Task {
	if t.waiting.count.Load() == 0 {
		return nil
	}

	t.waiting.mutex.Lock()
	defer t.waiting.mutex.Unlock()

	for _, group := range t.waiting.items {
		for _, task := range group.tasks {
			delete(group.tasks, task.Uuid)

			t.waiting.count.Add(-1)

			if len(group.tasks) == 0 {
				delete(t.waiting.items, group.uuid)
			}

			return task
		}
	}

	return nil
}

func (t *Tasks) AddFinished(task *Task) {
	addTaskToGroup(task, t.finished)
}

func (t *Tasks) TakeFinished(groupUuid string) *Task {
	if t.finished.count.Load() == 0 {
		return nil
	}

	t.finished.mutex.Lock()
	defer t.finished.mutex.Unlock()

	group, exists := t.finished.items[groupUuid]

	if !exists {
		return nil
	}

	for _, task := range group.tasks {
		delete(group.tasks, task.Uuid)

		t.finished.count.Add(-1)

		if len(group.tasks) == 0 {
			delete(t.finished.items, group.uuid)
		}

		return task
	}

	return nil
}

func (t *Tasks) FlushRottenFinished() {
	if t.finished.count.Load() == 0 {
		return
	}

	t.finished.mutex.Lock()
	defer t.finished.mutex.Unlock()

	for _, group := range t.finished.items {
		if !group.IsTimeout() {
			continue
		}

		delete(t.finished.items, group.uuid)

		break
	}
}

func (t *Tasks) Delete(task *Task) {
	deleteTaskFromGroup(task, t.waiting)
	deleteTaskFromGroup(task, t.finished)
}
