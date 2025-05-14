package commands

import "context"

type CommandInterface interface {
	Title() string
	Parameters() string
	Handle(ctx context.Context, arguments []string) error
	Close() error
}
