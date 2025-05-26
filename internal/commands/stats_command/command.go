package stats_command

import (
	"context"
	"os"
	"os/exec"
	"sparallel_server/internal/services/stats_service"
	"sync/atomic"
	"time"
)

type Command struct {
	closing atomic.Bool
}

func (c *Command) Title() string {
	return "Print stats"
}

func (c *Command) Parameters() string {
	return "{no parameters}"
}

func (c *Command) Handle(ctx context.Context, arguments []string) error {
	serv := stats_service.NewService()

	for !c.closing.Load() {
		time.Sleep(1 * time.Second)

		cmd := exec.Command("clear")

		cmd.Stdout = os.Stdout

		_ = cmd.Run()

		_ = serv.Print()
	}

	return nil
}

func (c *Command) Close() error {
	c.closing.Store(true)

	return nil
}
