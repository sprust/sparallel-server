package commands

import (
	"sparallel_server/internal/commands/hello_command"
	"sparallel_server/internal/commands/serve_rpc_command"
	"sparallel_server/internal/commands/stats_command"
	foundationCommands "sparallel_server/pkg/foundation/commands"
)

const (
	HelloCommandName    = "hello"
	ServeRpcCommandName = "start"
	StatsCommandName    = "stats"
)

var commands = map[string]foundationCommands.CommandInterface{
	HelloCommandName:    &hello_command.Command{},
	ServeRpcCommandName: &serve_rpc_command.Command{},
	StatsCommandName:    &stats_command.Command{},
}

func GetCommands() map[string]foundationCommands.CommandInterface {
	return commands
}
