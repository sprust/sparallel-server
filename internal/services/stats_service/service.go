package stats_service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"sparallel_server/internal/services/proxy_server/mongodb_proxy"
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/mongodb_proxy_objects"
	"sparallel_server/internal/services/workers_server"
	"time"
)

var statsFilePath = "storage/logs/stats.json"

var service *Service

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
	if service != nil {
		panic("stats service is already created")
	}

	service = &Service{}

	return service
}

func (s *Service) Save() {
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

	jsonData, err := json.Marshal(combined)

	if err != nil {
		slog.Error("Failed to marshal stats", "error", err)

		return
	}

	err = os.WriteFile(statsFilePath, jsonData, 0644)

	if err != nil {
		slog.Error("Failed to write stats file", "error", err)

		return
	}
}

func (s *Service) Print() error {
	content, err := os.ReadFile(statsFilePath)

	if err != nil {
		slog.Error("Failed to read stats file", "error", err)

		return err
	}

	var prettyJSON bytes.Buffer

	err = json.Indent(&prettyJSON, content, "", "    ")

	if err != nil {
		slog.Error("Failed to format JSON", "error", err)

		return err
	}

	fmt.Println(prettyJSON.String())

	return err
}
