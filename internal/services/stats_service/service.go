package stats_service

import (
	"runtime"
	"sparallel_server/internal/services/proxy_server/mongodb_proxy"
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/mongodb_proxy_objects"
	"sparallel_server/internal/services/workers_server"
	"time"
)

type SystemStats struct {
	NumGoroutine  uint64
	AllocMiB      float32
	TotalAllocMiB float32
	SysMiB        float32
	NumGC         uint64
}

type Service struct {
}

type CombinedStats struct {
	DateTime     time.Time                           `json:"dateTime"`
	System       SystemStats                         `json:"system"`
	Workers      *workers_server.WorkersServerStats  `json:"workers,omitempty"`
	MongodbProxy *mongodb_proxy_objects.ServiceStats `json:"mongodb_proxy,omitempty"`
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Get() CombinedStats {
	combined := CombinedStats{
		DateTime: time.Now(),
	}

	var mem runtime.MemStats

	runtime.ReadMemStats(&mem)

	sysStats := SystemStats{
		NumGoroutine:  uint64(runtime.NumGoroutine()),
		AllocMiB:      float32(mem.Alloc / 1024 / 1024),
		TotalAllocMiB: float32(mem.TotalAlloc / 1024 / 1024),
		SysMiB:        float32(mem.Sys / 1024 / 1024),
		NumGC:         uint64(mem.NumGC),
	}

	combined.System = sysStats

	workersService := workers_server.GetService()

	if workersService != nil {
		workersServiceStats := workersService.Stats()

		combined.Workers = &workersServiceStats
	}

	mongodbProxyService := mongodb_proxy.GetService()

	if mongodbProxyService != nil {
		mongodbProxyServiceStats := mongodbProxyService.Stats()

		combined.MongodbProxy = &mongodbProxyServiceStats
	}

	return combined
}
