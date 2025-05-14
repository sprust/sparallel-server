package atomic

import "sync/atomic"

type Boolean struct {
	value int32
}

func (b *Boolean) Set(boolVal bool) {
	var intVal int32

	if boolVal {
		intVal = 1
	} else {
		intVal = 0
	}

	atomic.StoreInt32(&b.value, intVal)
}

func (b *Boolean) Get() bool {
	return atomic.LoadInt32(&b.value) != 0
}
