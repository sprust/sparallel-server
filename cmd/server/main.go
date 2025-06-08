package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"sparallel_server/internal/commands"
	"sparallel_server/internal/config"
	"sparallel_server/pkg/foundation/app"
	appConfig "sparallel_server/pkg/foundation/config"
	"strconv"
	"strings"
)

var args = os.Args

func init() {
	env := flag.String("env", "", "Specify the environment file to load")

	flag.Parse()

	if env == nil || *env == "" {
		appConfig.Init()
	} else {
		appConfig.Init(*env)

		args = slices.DeleteFunc(args, func(arg string) bool {
			return strings.HasPrefix(arg, "--env=")
		})
	}
}

func main() {
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
		getAppConfig(),
		commands.GetCommands(),
		getServiceProviders(),
	)

	newApp.Start(commandName, commandArgs)
}

func getAppConfig() appConfig.Config {
	logKeepDays, err := strconv.Atoi(os.Getenv("LOG_KEEP_DAYS"))

	if err != nil {
		fmt.Println("LOG_KEEP_DAYS is not set or invalid, using default value 3")

		logKeepDays = 3
	}

	return appConfig.Config{
		LogConfig: appConfig.LogConfig{
			Levels:   getLogLevels(),
			DirPath:  os.Getenv("LOG_DIR"),
			KeepDays: logKeepDays,
		},
	}
}

func getLogLevels() []slog.Level {
	logLevels := strings.Split(config.GetConfig().GetLogLevels(), ",")

	var slogLevels []slog.Level

	if slices.Index(logLevels, "any") == -1 {
		for _, logLevel := range logLevels {
			if logLevel == "" {
				continue
			}

			switch logLevel {
			case "debug":
				slogLevels = append(slogLevels, slog.LevelDebug)
			case "info":
				slogLevels = append(slogLevels, slog.LevelInfo)
			case "warn":
				slogLevels = append(slogLevels, slog.LevelWarn)
			case "error":
				slogLevels = append(slogLevels, slog.LevelError)
			default:
				panic(fmt.Errorf("unknown log level: %s", logLevel))
			}
		}
	}

	return slogLevels
}

func getServiceProviders() []app.ServiceProviderInterface {
	return []app.ServiceProviderInterface{}
}
