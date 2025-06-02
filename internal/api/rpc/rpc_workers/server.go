package rpc_workers

import (
	"context"
	"log/slog"
	"sparallel_server/internal/config"
	"sparallel_server/internal/services/workers_server"
	"sync"
)

var server *WorkersServer
var once sync.Once

type WorkersServer struct {
	service *workers_server.Service
}

func NewServer(ctx context.Context) *WorkersServer {
	once.Do(func() {
		cfg := config.GetConfig()

		service := workers_server.NewService(
			cfg.GetCommand(),
			cfg.GetMinWorkersNumber(),
			cfg.GetMaxWorkersNumber(),
			cfg.GetWorkersNumberPercentScale(),
			cfg.GetWorkersNumberScaleUp(),
			cfg.GetWorkersNumberScaleDown(),
		)

		service.Start(ctx)

		server = &WorkersServer{
			service: service,
		}
	})

	return server
}

func (s *WorkersServer) AddTask(args *AddTaskArgs, reply *AddTaskResult) error {
	task := s.service.AddTask(args.GroupUuid, args.TaskUuid, args.UnixTimeout, args.Payload)

	reply.Uuid = task.TaskUuid

	return nil
}

func (s *WorkersServer) DetectAnyFinishedTask(args *DetectFinishedTaskArgs, reply *DetectFinishedTaskResult) error {
	response := s.service.DetectAnyFinishedTask(args.GroupUuid)

	reply.GroupUuid = response.GroupUuid
	reply.TaskUuid = response.TaskUuid
	reply.IsFinished = response.IsFinished
	reply.Response = response.Response
	reply.IsError = response.IsError

	return nil
}

func (s *WorkersServer) CancelGroup(args *CancelGroupArgs, reply *CancelGroupResult) error {
	s.service.CancelGroup(args.GroupUuid)

	reply.GroupUuid = args.GroupUuid

	return nil
}

func (s *WorkersServer) Close() error {
	slog.Warn("Closing workers server")

	return s.service.Close()
}
