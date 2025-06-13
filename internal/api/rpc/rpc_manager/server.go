package rpc_manager

import (
	"encoding/json"
	"log/slog"
	"sparallel_server/internal/services/stats_service"
	"sync"
	"syscall"
	"time"
)

var server *ManagerServer
var once sync.Once

type ManagerServer struct {
	service *stats_service.Service
}

func NewServer() *ManagerServer {
	once.Do(func() {
		server = &ManagerServer{
			service: stats_service.NewService(),
		}
	})

	return server
}

func (s *ManagerServer) Stop(args *StopArgs, reply *StopResult) error {
	slog.Warn("Stop server with message [" + args.Message + "]...")

	go func() {
		time.Sleep(1 * time.Second)

		_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}()

	reply.Answer = "Ok"

	return nil
}

func (s *ManagerServer) Stats(_ *StatsArgs, reply *StatsResult) error {
	stats := s.service.Get()

	jsonData, err := json.Marshal(stats)

	if err != nil {
		slog.Error("Failed to marshal stats: " + err.Error())

		return err
	}

	reply.Json = string(jsonData)

	return nil
}

func (s *ManagerServer) Close() error {
	slog.Warn("Closing manager server")

	return nil
}
