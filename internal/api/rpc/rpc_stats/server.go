package rpc_stats

import (
	"encoding/json"
	"log/slog"
	"sparallel_server/internal/services/stats_service"
)

type StatsServer struct {
	service *stats_service.Service
}

func NewServer() *StatsServer {
	return &StatsServer{
		service: stats_service.NewService(),
	}
}

type StatsArgs struct {
	Message string
}

type GetResult struct {
	Json string
}

func (s *StatsServer) Get(args *StatsArgs, reply *GetResult) error {
	slog.Debug("Get stats with message [" + args.Message + "]")

	stats := s.service.Get()

	jsonData, err := json.Marshal(stats)

	if err != nil {
		slog.Error("Failed to marshal stats: " + err.Error())

		return err
	}

	reply.Json = string(jsonData)

	return nil
}
func (s *StatsServer) Close() error {
	slog.Warn("Closing stats server")

	return nil
}
