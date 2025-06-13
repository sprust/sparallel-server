package commands

import (
	"context"
	"io"
	"sparallel_server/pkg/foundation/app_io"
)

type CommandInterface interface {
	Title() string
	Parameters() string
	Handle(ctx context.Context, arguments []string) error
	app_io.Pauser
	io.Closer
}
