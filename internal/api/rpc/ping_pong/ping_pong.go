package ping_pong

import (
	"log/slog"
	"strconv"
)

type PingPong struct {
}

type PingPongArgs struct {
	Message string
}

type PingPongResult struct {
	Message string
}

func (p *PingPong) Ping(args *PingPongArgs, reply *PingPongResult) error {
	reply.Message = args.Message

	go slog.Info("Ping: " + strconv.Itoa(len(args.Message)))

	return nil
}
