package sparallel

import (
	"fmt"
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

func (s *Server) AddTask(args *AddTaskArgs, reply *AddTaskResult) error {
	task := s.SparallelServer.AddTask(args.GroupUuid, args.UnixTimeout, args.Payload)

	fmt.Println(task)

	reply.Uuid = task.Uuid

	return nil
}
