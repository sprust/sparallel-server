package main

import (
	"flag"
	"os"
	"slices"
	"sparallel_server/internal/commands"
	pConfig "sparallel_server/internal/config"
	"sparallel_server/pkg/foundation/app"
	fConfig "sparallel_server/pkg/foundation/config"
	"strings"
)

var args = os.Args

func init() {
	env := flag.String("env", "", "Specify the environment file to load")

	flag.Parse()

	if env == nil || *env == "" {
		fConfig.Init()
	} else {
		fConfig.Init(*env)

		args = slices.DeleteFunc(args, func(arg string) bool {
			return strings.HasPrefix(arg, "--env=")
		})
	}
}

func main() {
	fCfg := fConfig.GetConfig()

	pConfig.Init(fCfg)

	var commandName string
	var commandArgs []string

	if len(args) > 1 {
		commandName = args[1]
	}

	if len(args) > 2 {
		argsSlice := args[2:]

		commandArgs = make([]string, len(argsSlice))

		copy(commandArgs, argsSlice)
	}

	newApp := app.NewApp(
		fCfg,
		commands.GetCommands(),
		getServiceProviders(),
	)

	newApp.Start(commandName, commandArgs)
}

func getServiceProviders() []app.ServiceProviderInterface {
	return []app.ServiceProviderInterface{}
}
