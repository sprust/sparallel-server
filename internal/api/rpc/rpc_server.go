package rpc

import (
	"context"
	"errors"
	"fmt"
	goridgeRpc "github.com/roadrunner-server/goridge/v3/pkg/rpc"
	"log/slog"
	"net"
	"net/rpc"
	"os"
	"sparallel_server/internal/api/rpc/rpc_manager"
	"sparallel_server/internal/api/rpc/rpc_ping_pong"
	"sparallel_server/internal/api/rpc/rpc_proxy_mongodb"
	"sparallel_server/internal/api/rpc/rpc_workers"
	"sparallel_server/internal/config"
	"sparallel_server/pkg/foundation/errs"
	"sync"
	"sync/atomic"
)

var server *Server
var once sync.Once

type Server struct {
	rpcPort      string
	listener     net.Listener
	servers      []ServerInterface
	pausingMutex sync.Mutex
	closing      atomic.Bool
	closed       atomic.Bool
	config       *config.Config
}

func NewServer(rpcPort string) *Server {
	once.Do(func() {
		server = &Server{
			rpcPort: rpcPort,
			config:  config.GetConfig(),
		}
	})

	return server
}

func (s *Server) Run(ctx context.Context) error {
	listener, err := net.Listen("tcp", ":"+s.rpcPort)

	if err != nil {
		return errs.Err(err)
	}

	s.listener = listener

	for _, srv := range s.detectServers(ctx) {
		err = rpc.Register(srv)

		if err != nil {
			slog.Error(err.Error())

			return errs.Err(err)
		}

		s.servers = append(s.servers, srv)
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
			slog.Error("Error listening: " + err.Error())

			continue
		}

		go rpc.ServeCodec(goridgeRpc.NewCodec(conn))
	}

	for !s.closed.Load() {
		//
	}

	return nil
}

func (s *Server) GetServers() []ServerInterface {
	return s.servers
}

func (s *Server) Pause() error {
	if !s.pausingMutex.TryLock() {
		return errs.Err(errors.New("server is already paused"))
	}

	defer s.pausingMutex.Unlock()

	for _, srv := range s.servers {
		err := srv.Pause()

		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) UnPause() error {
	if !s.pausingMutex.TryLock() {
		return errs.Err(errors.New("server is pausing"))
	}

	defer s.pausingMutex.Unlock()

	for _, srv := range s.servers {
		err := srv.UnPause()

		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) Close() error {
	slog.Warn("Closing rpc server...")

	s.closing.Store(true)

	var errList []error

	for _, server := range s.servers {
		err := server.Close()

		if err != nil {
			errList = append(errList, err)
		}
	}

	if s.listener != nil {
		err := errs.Err(s.listener.Close())

		if err != nil {
			errList = append(errList, err)
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

	return errs.Err(joinErrors(errList))
}

func (s *Server) detectServers(ctx context.Context) []ServerInterface {
	servers := []ServerInterface{
		rpc_ping_pong.NewServer(),
		rpc_manager.NewServer(),
	}

	if s.config.IsServeWorkers() {
		servers = append(servers, rpc_workers.NewServer(ctx))
	}

	if s.config.IsServeProxy() {
		servers = append(servers, rpc_proxy_mongodb.NewServer(ctx))
	}

	return servers
}

func joinErrors(errList []error) error {
	if len(errList) == 0 {
		return nil
	}

	errorMessage := ""

	for i, err := range errList {
		if i > 0 {
			errorMessage += "; "
		}

		errorMessage += err.Error()
	}

	return errors.New(errorMessage)
}
