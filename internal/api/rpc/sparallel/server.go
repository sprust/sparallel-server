package sparallel

import (
	"sparallel_server/internal/services/sparallel_server"
)

type Server struct {
	SparallelServer *sparallel_server.Server
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
	task := s.SparallelServer.AddTask(args.GroupUuid, args.TaskUuid, args.UnixTimeout, args.Payload)

	reply.Uuid = task.TaskUuid

	return nil
}

func (s *Server) DetectAnyFinishedTask(args *DetectFinishedTaskArgs, reply *DetectFinishedTaskResult) error {
	response := s.SparallelServer.DetectAnyFinishedTask(args.GroupUuid)

	reply.GroupUuid = response.GroupUuid
	reply.TaskUuid = response.TaskUuid
	reply.IsFinished = response.IsFinished
	reply.Response = response.Response
	reply.IsError = response.IsError

	return nil
}

func (s *Server) CancelGroup(args *CancelGroupArgs, reply *CancelGroupResult) error {
	s.SparallelServer.CancelGroup(args.GroupUuid)

	reply.GroupUuid = args.GroupUuid

	return nil
}
