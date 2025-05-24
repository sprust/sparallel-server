package insert_one

import "sync"

type Operations struct {
	mutex sync.Mutex
	items map[string]*Operation
}

func NewOperations() *Operations {
	return &Operations{
		items: make(map[string]*Operation),
	}
}

func (l *Operations) Add(action *Operation) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.items[action.Uuid()] = action
}

func (l *Operations) Pull(uuid string) *Operation {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	operation, exists := l.items[uuid]

	if !exists {
		return nil
	}

	delete(l.items, uuid)

	return operation
}
