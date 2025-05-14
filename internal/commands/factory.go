package commands

import (
	"sparallel_server/internal/commands/hello_command"
	foundationCommands "sparallel_server/pkg/foundation/commands"
)

const (
	HelloCommandName = "hello"
)

var commands = map[string]foundationCommands.CommandInterface{
	HelloCommandName: &hello_command.Command{},
}

func GetCommands() map[string]foundationCommands.CommandInterface {
	return commands
}
