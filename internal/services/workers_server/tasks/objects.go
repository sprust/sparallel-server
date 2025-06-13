package tasks

import "sync/atomic"

type Tasks struct {
	waiting  *SubTasks
	finished *SubTasks

	addedTotalCount    atomic.Int64
	reAddedTotalCount  atomic.Int64
	tookTotalCount     atomic.Int64
	finishedTotalCount atomic.Int64
	successTotalCount  atomic.Int64
	errorTotalCount    atomic.Int64
	timeoutTotalCount  atomic.Int64
}

type Group struct {
	uuid        string
	unixTimeout int
	tasks       map[string]*Task // map[TaskUuid]
}

func (g *Group) IsTimeout() bool {
	return isTimeout(g.unixTimeout, 5)
}

type Task struct {
	GroupUuid   string
	TaskUuid    string
	UnixTimeout int
	Payload     string
	IsFinished  bool
	Response    string
	IsError     bool
}

func (t *Task) IsTimeout() bool {
	return isTimeout(t.UnixTimeout, 5)
}

type OrderedGroups struct {
	data  map[string]*Group
	order []string
}

func NewOrderedGroups() *OrderedGroups {
	return &OrderedGroups{
		data:  make(map[string]*Group),
		order: make([]string, 0),
	}
}

func (m *OrderedGroups) Add(groupUuid string, group *Group) {
	if _, exists := m.data[groupUuid]; !exists {
		m.order = append(m.order, groupUuid)
	}
	m.data[groupUuid] = group
}

func (m *OrderedGroups) Delete(groupUuid string) {
	delete(m.data, groupUuid)
	for i, uuid := range m.order {
		if uuid == groupUuid {
			m.order = append(m.order[:i], m.order[i+1:]...)
			break
		}
	}
}
