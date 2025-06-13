package workers

import "sync/atomic"

func (w *Workers) GetCount() int {
	return int(w.totalCount.Load())
}

func (w *Workers) GetBusyCount() int {
	return int(w.busyCount.Load())
}

func (w *Workers) GetFreeCount() int {
	return int(w.freeCount.Load())
}

func (w *Workers) GetLoadPercent() int {
	count := w.GetCount()

	if count == 0 {
		return 0
	}

	return w.GetBusyCount() * 100 / count
}

func (w *Workers) GetAddedCount() int {
	return int(w.addedCount.Load())
}

func (w *Workers) GetTookCount() int {
	return int(w.tookCount.Load())
}

func (w *Workers) GetFreedCount() int {
	return int(w.freedCount.Load())
}

func (w *Workers) GetDeletedCount() int {
	return int(w.deletedCount.Load())
}

func (w *Workers) incCount(counter *atomic.Int64) {
	go counter.Add(1)
}
