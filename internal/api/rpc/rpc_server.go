package rpc

import (
	"context"
	"fmt"
	goridgeRpc "github.com/roadrunner-server/goridge/v3/pkg/rpc"
	"io"
	"log/slog"
	"net"
	"net/rpc"
	"os"
	"sparallel_server/internal/api/rpc/rpc_ping_pong"
	"sparallel_server/internal/api/rpc/rpc_proxy_mongodb"
	"sparallel_server/internal/api/rpc/rpc_stats"
	"sparallel_server/internal/api/rpc/rpc_workers"
	"sparallel_server/internal/config"
	"sparallel_server/pkg/foundation/errs"
	"sync/atomic"
	"time"
)

type Server struct {
	rpcPort  string
	listener net.Listener
	servers  []io.Closer
	ticker   *time.Ticker
	closing  atomic.Bool
	closed   atomic.Bool
	config   *config.Config
}

func NewServer(rpcPort string) *Server {
	server := &Server{
		rpcPort: rpcPort,
		config:  config.GetConfig(),
	}

	return server
}

func (s *Server) Run(ctx context.Context) error {
	listener, err := net.Listen("tcp", ":"+s.rpcPort)

	if err != nil {
		return errs.Err(err)
	}

	s.listener = listener

	for _, server := range s.getServers(ctx) {
		err = rpc.Register(server)

		if err != nil {
			slog.Error(err.Error())

			return errs.Err(err)
		}

		s.servers = append(s.servers, server)
	}

	pidFilePath := s.config.GetServerPidFilePath()

	if pidFilePath != "" {
		pid := os.Getpid()
		err = os.WriteFile(pidFilePath, []byte(fmt.Sprint(pid)), 0644)
		if err != nil {
			return errs.Err(err)
		}

		slog.Warn("Pid file created: " + pidFilePath)
	}

	slog.Info("Listening on port " + s.rpcPort)

	for {
		conn, err := s.listener.Accept()

		if s.closing.Load() {
			break
		}

		if err != nil {
			slog.Error("Error listening:", err.Error())

			continue
		}

		_ = conn

		go rpc.ServeCodec(goridgeRpc.NewCodec(conn))
	}

	for !s.closed.Load() {
		//
	}

	return nil
}

func (s *Server) Close() error {
	slog.Warn("Closing rpc server...")

	s.closing.Store(true)

	var errors []error

	for _, server := range s.servers {
		err := server.Close()

		if err != nil {
			errors = append(errors, err)
		}
	}

	if s.listener != nil {
		err := errs.Err(s.listener.Close())

		if err != nil {
			errors = append(errors, err)
		}
	}

	pidFilePath := s.config.GetServerPidFilePath()

	if pidFilePath != "" {
		_, err := os.Stat(pidFilePath)

		if err == nil {
			_ = os.Remove(pidFilePath)

			slog.Warn("Pid file deleted: " + pidFilePath)
		}
	}

	s.closed.Store(true)

	return errs.Err(joinErrors(errors))
}

func (s *Server) getServers(ctx context.Context) []io.Closer {
	servers := []io.Closer{
		rpc_ping_pong.NewServer(),
		rpc_stats.NewServer(),
	}

	if s.config.IsServeWorkers() {
		servers = append(servers, rpc_workers.NewServer(ctx))
	}

	if s.config.IsServeProxy() {
		servers = append(servers, rpc_proxy_mongodb.NewServer(ctx))
	}

	return servers
}

func joinErrors(errors []error) error {
	if len(errors) == 0 {
		return nil
	}

	errorMessage := ""

	for i, err := range errors {
		if i > 0 {
			errorMessage += "; "
		}

		errorMessage += err.Error()
	}

	return fmt.Errorf(errorMessage)
}
