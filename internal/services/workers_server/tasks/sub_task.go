package tasks

import (
	"sync"
)

type SubTasks struct {
	mutex  sync.Mutex
	groups *OrderedGroups
}

func NewSubTasks() *SubTasks {
	return &SubTasks{
		mutex:  sync.Mutex{},
		groups: NewOrderedGroups(),
	}
}

func (s *SubTasks) AddTask(task *Task) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	group, exists := s.groups.data[task.GroupUuid]

	if !exists {
		group = &Group{
			uuid:        task.GroupUuid,
			unixTimeout: task.UnixTimeout,
			tasks:       make(map[string]*Task),
		}
		s.groups.Add(task.GroupUuid, group)
	}

	group.tasks[task.TaskUuid] = task
}

func (s *SubTasks) DeleteGroup(groupUuid string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.groups.Delete(groupUuid)
}

func (s *SubTasks) DeleteTask(task *Task) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	group, exists := s.groups.data[task.GroupUuid]

	if !exists {
		return
	}

	delete(group.tasks, task.TaskUuid)

	if len(group.tasks) == 0 {
		s.groups.Delete(task.GroupUuid)
	}
}

func (s *SubTasks) Pop() *Task {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Итерация по группам в порядке их добавления
	for _, groupUuid := range s.groups.order {
		group := s.groups.data[groupUuid]
		for taskUuid, task := range group.tasks {
			delete(group.tasks, taskUuid)

			if len(group.tasks) == 0 {
				s.groups.Delete(group.uuid)
			}

			return task
		}
	}

	return nil
}

func (s *SubTasks) TakeFirstByGroupUuid(groupUuid string) *Task {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	group, exists := s.groups.data[groupUuid]

	if !exists {
		return nil
	}

	for taskUuid, task := range group.tasks {
		delete(group.tasks, taskUuid)

		if len(group.tasks) == 0 {
			s.groups.Delete(group.uuid)
		}

		return task
	}

	return nil
}

func (s *SubTasks) FlushFirstRotten() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, groupUuid := range s.groups.order {
		group := s.groups.data[groupUuid]
		if !group.IsTimeout() {
			continue
		}

		itemsCount := len(group.tasks)
		s.groups.Delete(group.uuid)
		return itemsCount
	}

	return 0
}

func (s *SubTasks) GetCount() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var count int

	for _, group := range s.groups.data {
		count += len(group.tasks)
	}

	return count
}
