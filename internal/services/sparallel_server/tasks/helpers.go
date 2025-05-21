package tasks

import "time"

func addTaskToGroup(task *Task, subTasks *SubTasks) {
	subTasks.mutex.Lock()
	defer subTasks.mutex.Unlock()

	group, exists := subTasks.items[task.GroupUuid]

	if !exists {
		group = &Group{
			uuid:        task.GroupUuid,
			unixTimeout: task.UnixTimeout,
			tasks:       make(map[string]*Task),
		}
	}

	group.tasks[task.Uuid] = task

	subTasks.count.Add(1)
}

func deleteTaskFromGroup(task *Task, subTasks *SubTasks) bool {
	subTasks.mutex.Lock()
	defer subTasks.mutex.Unlock()

	group, exists := subTasks.items[task.GroupUuid]

	if !exists {
		return false
	}

	delete(group.tasks, task.Uuid)

	if len(group.tasks) == 0 {
		delete(subTasks.items, task.GroupUuid)
	}

	subTasks.count.Add(-1)

	return true
}

func isTimeout(unixTimeout int, headStart int) bool {
	now := time.Now().Unix()

	return (int64(unixTimeout) - now) < -int64(headStart)
}
