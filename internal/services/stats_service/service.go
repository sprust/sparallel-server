package stats_service

import (
	"runtime"
	"sparallel_server/internal/services/workers_service"
	"sync"
	"time"
)

var service *Service
var once sync.Once

type SystemStats struct {
	NumGoroutine  uint64
	AllocMiB      float64
	TotalAllocMiB float64
	SysMiB        float64
	NumGC         uint64
}

type Service struct {
}

type CombinedStats struct {
	DateTime time.Time                           `json:"dateTime"`
	System   SystemStats                         `json:"system"`
	Workers  *workers_service.WorkersServerStats `json:"workers,omitempty"`
}

func NewService() *Service {
	once.Do(func() {
		service = &Service{}
	})

	return service
}

func (s *Service) Get() CombinedStats {
	combined := CombinedStats{
		DateTime: time.Now(),
	}

	var mem runtime.MemStats

	runtime.ReadMemStats(&mem)

	sysStats := SystemStats{
		NumGoroutine:  uint64(runtime.NumGoroutine()),
		AllocMiB:      float64(mem.Alloc / 1024 / 1024),
		TotalAllocMiB: float64(mem.TotalAlloc / 1024 / 1024),
		SysMiB:        float64(mem.Sys / 1024 / 1024),
		NumGC:         uint64(mem.NumGC),
	}

	combined.System = sysStats

	workersService := workers_service.GetService()

	if workersService != nil {
		workersServiceStats := workersService.Stats()

		combined.Workers = &workersServiceStats
	}

	return combined
}
