package serve_rpc_command

import (
	"context"
	"sparallel_server/internal/api/rpc"
	"sparallel_server/internal/config"
	"sparallel_server/pkg/foundation/errs"
)

type Command struct {
	server *rpc.Server
}

func (c *Command) Title() string {
	return "Serve rpc"
}

func (c *Command) Parameters() string {
	return "{no parameters}"
}

func (c *Command) Handle(ctx context.Context, arguments []string) error {
	rpcPort := config.GetConfig().GetRpcPort()

	c.server = rpc.NewServer(rpcPort)

	err := c.server.Run(ctx)

	return errs.Err(err)
}

func (c *Command) Close() error {
	if c.server != nil {
		return errs.Err(c.server.Close())
	}

	return nil
}
