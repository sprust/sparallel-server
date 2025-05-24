package workers

import (
	"context"
	"sparallel_server/internal/config"
	"sparallel_server/internal/services/sparallel_server"
)

var workers *Server

type Server struct {
	service *sparallel_server.Service
}

func NewWorkersServer(ctx context.Context) *Server {
	if workers != nil {
		panic("the workers is already created")
	}

	cfg := config.GetConfig()

	service := sparallel_server.NewService(
		cfg.GetCommand(),
		cfg.GetMinWorkersNumber(),
		cfg.GetMaxWorkersNumber(),
		cfg.GetWorkersNumberPercentScale(),
		cfg.GetWorkersNumberScaleUp(),
		cfg.GetWorkersNumberScaleDown(),
	)

	service.Start(ctx)

	server := &Server{
		service: service,
	}

	return server
}

type AddTaskArgs struct {
	GroupUuid   string
	TaskUuid    string
	UnixTimeout int
	Payload     string
}

type AddTaskResult struct {
	Uuid string
}

type DetectFinishedTaskArgs struct {
	GroupUuid string
}

type DetectFinishedTaskResult struct {
	GroupUuid  string
	TaskUuid   string
	IsFinished bool
	Response   string
	IsError    bool
}

type CancelGroupArgs struct {
	GroupUuid string
}

type CancelGroupResult struct {
	GroupUuid string
}

func (s *Server) AddTask(args *AddTaskArgs, reply *AddTaskResult) error {
	task := s.service.AddTask(args.GroupUuid, args.TaskUuid, args.UnixTimeout, args.Payload)

	reply.Uuid = task.TaskUuid

	return nil
}

func (s *Server) DetectAnyFinishedTask(args *DetectFinishedTaskArgs, reply *DetectFinishedTaskResult) error {
	response := s.service.DetectAnyFinishedTask(args.GroupUuid)

	reply.GroupUuid = response.GroupUuid
	reply.TaskUuid = response.TaskUuid
	reply.IsFinished = response.IsFinished
	reply.Response = response.Response
	reply.IsError = response.IsError

	return nil
}

func (s *Server) CancelGroup(args *CancelGroupArgs, reply *CancelGroupResult) error {
	s.service.CancelGroup(args.GroupUuid)

	reply.GroupUuid = args.GroupUuid

	return nil
}
