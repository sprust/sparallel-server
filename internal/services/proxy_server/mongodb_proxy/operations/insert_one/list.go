package insert_one

import "sync"

type List struct {
	mutex sync.Mutex
	items map[string]*Operation
}

func NewList() *List {
	return &List{}
}

func (l *List) Add(action *Operation) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.items[action.Uuid()] = action
}

func (l *List) Get(uuid string) *Operation {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return l.items[uuid]
}
