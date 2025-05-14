package hello_command

import (
	"context"
	"fmt"
)

type Command struct {
}

func (c *Command) Title() string {
	return "Just print Hello"
}

func (c *Command) Parameters() string {
	return ""
}

func (c *Command) Handle(ctx context.Context, arguments []string) error {
	fmt.Println("hello")

	return nil
}

func (c *Command) Close() error {
	return nil
}
