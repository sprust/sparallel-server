package rpc_ping_pong

import (
	"log/slog"
	"strconv"
)

type PingPongServer struct {
}

func NewServer() *PingPongServer {
	return &PingPongServer{}
}

type PingArgs struct {
	Message string
}

type PingResult struct {
	Message string
}

func (s *PingPongServer) Ping(args *PingArgs, reply *PingResult) error {
	reply.Message = args.Message

	go slog.Info("Ping: " + strconv.Itoa(len(args.Message)))

	return nil
}

func (s *PingPongServer) Pause() error {
	return nil
}

func (s *PingPongServer) UnPause() error {
	return nil
}

func (s *PingPongServer) Close() error {
	slog.Warn("Closing ping-pong server")

	return nil
}
