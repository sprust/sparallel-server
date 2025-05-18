package sparallel

import (
	"sparallel_server/internal/services/sparallel_server"
)

type Server struct {
	SparallelServer *sparallel_server.Server
}

type AddTaskArgs struct {
	GroupUuid   string
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
	IsFinished bool
	Response   string
	IsError    bool
}

func (s *Server) AddTask(args *AddTaskArgs, reply *AddTaskResult) error {
	task := s.SparallelServer.AddTask(args.GroupUuid, args.UnixTimeout, args.Payload)

	reply.Uuid = task.Uuid

	return nil
}

func (s *Server) DetectAnyFinishedTask(args *DetectFinishedTaskArgs, reply *DetectFinishedTaskResult) error {
	response := s.SparallelServer.DetectAnyFinishedTask(args.GroupUuid)

	reply.IsFinished = response.IsFinished
	reply.Response = response.Response
	reply.IsError = response.IsError

	return nil
}
