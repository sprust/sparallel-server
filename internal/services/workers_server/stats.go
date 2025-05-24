package workers_server

type SystemStats struct {
	NumGoroutine  uint64
	AllocMiB      float32
	TotalAllocMiB float32
	SysMiB        float32
	NumGC         uint64
}
