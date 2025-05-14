package atomic

import (
	"sync/atomic"
)

type Counter struct {
	value int32
}

func (c *Counter) Increment() {
	atomic.AddInt32(&c.value, 1)
}

func (c *Counter) Decrement() {
	atomic.AddInt32(&c.value, -1)
}

func (c *Counter) Get() int32 {
	return atomic.LoadInt32(&c.value)
}
