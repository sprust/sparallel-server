package helpers

import (
	"sync/atomic"
)

func IncInt64Async(counter *atomic.Int64) {
	IncInt64AsyncDelta(counter, 1)
}

func IncInt64AsyncDelta(counter *atomic.Int64, delta int) {
	go counter.Add(int64(delta))
}
