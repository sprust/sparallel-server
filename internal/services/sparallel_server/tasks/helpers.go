package tasks

import "time"

func addTaskToGroup(task *Task, subTasks *SubTasks) {
	subTasks.mutex.Lock()
	defer subTasks.mutex.Unlock()

	group, exists := subTasks.groups[task.GroupUuid]

	if !exists {
		group = &Group{
			uuid:        task.GroupUuid,
			unixTimeout: task.UnixTimeout,
			tasks:       make(map[string]*Task),
		}

		subTasks.groups[task.GroupUuid] = group
	}

	group.tasks[task.Uuid] = task

	subTasks.count.Add(1)
}

func deleteTaskFromGroup(task *Task, subTasks *SubTasks) bool {
	subTasks.mutex.Lock()
	defer subTasks.mutex.Unlock()

	group, exists := subTasks.groups[task.GroupUuid]

	if !exists {
		return false
	}

	delete(group.tasks, task.Uuid)

	if len(group.tasks) == 0 {
		delete(subTasks.groups, task.GroupUuid)
	}

	subTasks.count.Add(-1)

	return true
}

func flushFirstRotten(subTasks *SubTasks) {
	if subTasks.count.Load() == 0 {
		return
	}

	subTasks.mutex.Lock()
	defer subTasks.mutex.Unlock()

	for _, group := range subTasks.groups {
		if !group.IsTimeout() {
			continue
		}

		itemsCount := len(group.tasks)

		delete(subTasks.groups, group.uuid)

		if itemsCount > 0 {
			subTasks.count.Add(-int32(itemsCount))
		}

		break
	}
}

func isTimeout(unixTimeout int, headStart int) bool {
	now := time.Now().Unix()

	return (int64(unixTimeout) - now) < -int64(headStart)
}
