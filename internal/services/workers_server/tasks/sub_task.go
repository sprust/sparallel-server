package tasks

import (
	"sync"
)

// TODO: FIFO for groups

type SubTasks struct {
	mutex  sync.Mutex
	groups map[string]*Group // map[GroupUuid]
}

func NewSubTasks() *SubTasks {
	return &SubTasks{
		mutex:  sync.Mutex{},
		groups: make(map[string]*Group),
	}
}

func (s *SubTasks) AddTask(task *Task) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	group, exists := s.groups[task.GroupUuid]

	if !exists {
		group = &Group{
			uuid:        task.GroupUuid,
			unixTimeout: task.UnixTimeout,
			tasks:       make(map[string]*Task),
		}

		s.groups[task.GroupUuid] = group
	}

	group.tasks[task.TaskUuid] = task
}

func (s *SubTasks) DeleteGroup(groupUuid string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.groups, groupUuid)
}

func (s *SubTasks) DeleteTask(task *Task) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	group, exists := s.groups[task.GroupUuid]

	if !exists {
		return
	}

	delete(group.tasks, task.TaskUuid)

	if len(group.tasks) == 0 {
		delete(s.groups, task.GroupUuid)
	}
}

func (s *SubTasks) Pop() *Task {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, group := range s.groups {
		for _, task := range group.tasks {
			delete(group.tasks, task.TaskUuid)

			if len(group.tasks) == 0 {
				delete(s.groups, group.uuid)
			}

			return task
		}
	}

	return nil
}

func (s *SubTasks) TakeFirstByGroupUuid(groupUuid string) *Task {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	group, exists := s.groups[groupUuid]

	if !exists {
		return nil
	}

	for _, task := range group.tasks {
		delete(group.tasks, task.TaskUuid)

		if len(group.tasks) == 0 {
			delete(s.groups, group.uuid)
		}

		return task
	}

	return nil
}

func (s *SubTasks) FlushFirstRotten() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, group := range s.groups {
		if !group.IsTimeout() {
			continue
		}

		itemsCount := len(group.tasks)

		delete(s.groups, group.uuid)

		return itemsCount
	}

	return 0
}

func (s *SubTasks) Count() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var count int

	for _, group := range s.groups {
		count += len(group.tasks)
	}

	return count
}
