package tasks

import (
	"sync"
	"sync/atomic"
)

type Tasks struct {
	waiting  *SubTasks
	finished *SubTasks
}

type SubTasks struct {
	mutex  sync.Mutex
	groups map[string]*Group // map[GroupUuid]
	count  atomic.Int32
}

func NewSubTasks() *SubTasks {
	return &SubTasks{
		mutex:  sync.Mutex{},
		groups: make(map[string]*Group),
		count:  atomic.Int32{},
	}
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
	Uuid        string
	UnixTimeout int
	Payload     string
	IsFinished  bool
	Response    string
	IsError     bool
}

func (t *Task) IsTimeout() bool {
	return isTimeout(t.UnixTimeout, 5)
}
