package rpc

import (
	"context"
	goridgeRpc "github.com/roadrunner-server/goridge/v3/pkg/rpc"
	"log/slog"
	"net"
	"net/rpc"
	"sparallel_server/internal/api/rpc/ping_pong"
	"sparallel_server/internal/api/rpc/proxy_mongodb"
	"sparallel_server/internal/api/rpc/workers"
	"sparallel_server/internal/services/sparallel_server"
	"sparallel_server/pkg/foundation/errs"
)

type Server struct {
	rpcPort         string
	listener        net.Listener
	sparallelServer *sparallel_server.Service
	closing         bool
}

func NewServer(rpcPort string) *Server {
	server := &Server{
		rpcPort: rpcPort,
	}

	return server
}

func (s *Server) Run(ctx context.Context) error {

	listener, err := net.Listen("tcp", ":"+s.rpcPort)

	if err != nil {
		return errs.Err(err)
	}

	s.listener = listener

	for _, function := range s.getFunctions(ctx) {
		err = rpc.Register(function)

		if err != nil {
			slog.Error(err.Error())

			return errs.Err(err)
		}
	}

	slog.Info("Listening on port " + s.rpcPort)

	for {
		conn, err := s.listener.Accept()

		if s.closing == true {
			break
		}

		if err != nil {
			slog.Error("Error listening:", err.Error())

			continue
		}

		_ = conn

		go rpc.ServeCodec(goridgeRpc.NewCodec(conn))
	}

	return nil
}

func (s *Server) getFunctions(ctx context.Context) []any {
	return []any{
		&ping_pong.PingPong{},
		workers.NewWorkersServer(ctx),
		proxy_mongodb.NewMongodbProxyServer(ctx),
	}
}

func (s *Server) Close() error {
	slog.Warn("Closing rpc server...")

	s.closing = true

	err := s.sparallelServer.Close()

	if err != nil {
		return err
	}

	return errs.Err(s.listener.Close())
}
