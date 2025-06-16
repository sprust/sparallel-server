package rpc_manager

import (
	"encoding/json"
	"log/slog"
	"sparallel_server/internal/services/stats_service"
	"syscall"
	"time"
)

type ManagerServer struct {
	service *stats_service.Service
}

func NewServer() *ManagerServer {
	return &ManagerServer{
		service: stats_service.NewService(),
	}
}

func (s *ManagerServer) Sleep(args *SleepArgs, reply *SleepResult) error {
	slog.Warn("Sleep server with message [" + args.Message + "]...")

	go func() {
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGTSTP)
	}()

	reply.Answer = "Ok"

	return nil
}

func (s *ManagerServer) WakeUp(args *WakeUpArgs, reply *WakeUpResult) error {
	slog.Warn("Wake up server with message [" + args.Message + "]...")

	go func() {
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGCONT)
	}()

	reply.Answer = "Ok"

	return nil
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

func (s *ManagerServer) Pause() error {
	return nil
}

func (s *ManagerServer) UnPause() error {
	return nil
}

func (s *ManagerServer) Close() error {
	slog.Warn("Closing manager server")

	return nil
}
