package rpc

import (
	"context"
	"fmt"
	goridgeRpc "github.com/roadrunner-server/goridge/v3/pkg/rpc"
	"io"
	"log/slog"
	"net"
	"net/rpc"
	"sparallel_server/internal/api/rpc/rpc_ping_pong"
	"sparallel_server/internal/api/rpc/rpc_proxy_mongodb"
	"sparallel_server/internal/api/rpc/rpc_workers"
	"sparallel_server/internal/config"
	"sparallel_server/internal/services/stats_service"
	"sparallel_server/pkg/foundation/errs"
	"time"
)

type Server struct {
	rpcPort  string
	listener net.Listener
	servers  []io.Closer
	ticker   *time.Ticker
	closing  bool
}

func NewServer(rpcPort string) *Server {
	server := &Server{
		rpcPort: rpcPort,
		ticker:  time.NewTicker(1 * time.Second),
	}

	return server
}

func (s *Server) Run(ctx context.Context) error {
	statsService := stats_service.NewService()

	go func(_ context.Context, statsService *stats_service.Service) {
		defer s.ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-s.ticker.C:
				statsService.Save()
			}
		}
	}(ctx, statsService)

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

func (s *Server) Close() error {
	slog.Warn("Closing rpc server...")

	s.closing = true

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

	return errs.Err(joinErrors(errors))
}

func (s *Server) getServers(ctx context.Context) []io.Closer {
	servers := []io.Closer{
		rpc_ping_pong.NewServer(),
	}

	cfg := config.GetConfig()

	if cfg.IsServeWorkers() {
		servers = append(servers, rpc_workers.NewServer(ctx))
	}

	if cfg.IsServeProxy() {
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
