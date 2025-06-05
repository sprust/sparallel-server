package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"sparallel_server/pkg/foundation/commands"
	"sparallel_server/pkg/foundation/config"
	"sparallel_server/pkg/foundation/errs"
	"sparallel_server/pkg/foundation/logging"
	"strings"
	"syscall"
)

type App struct {
	config             config.Config
	commands           map[string]commands.CommandInterface
	serviceProviders   []ServiceProviderInterface
	closeListeners     []io.Closer
	lastCloseListeners []io.Closer
}

func NewApp(
	config config.Config,
	commands map[string]commands.CommandInterface,
	serviceProviders []ServiceProviderInterface,
) App {
	app := App{
		config:           config,
		commands:         commands,
		serviceProviders: serviceProviders,
	}

	return app
}

func (a *App) Start(commandName string, args []string) {
	defer func(a *App) {
		if r := recover(); r != nil {
			err := a.Close()

			if err != nil {
				panic(errs.Err(err))
			}
		}
	}(a)

	if commandName == "" {
		fmt.Println("Commands:")

		for key, command := range a.commands {
			fmt.Printf(" %s %s - %s\n", key, command.Parameters(), command.Title())
		}

		return
	}

	a.initLogging()

	command, ok := a.commands[commandName]

	if !ok {
		panic(errs.Err(errors.New("command not found")))
	}

	a.AddFirstCloseListener(command)

	for _, provider := range a.serviceProviders {
		err := provider.Register()

		if err != nil {
			panic(err)
		}
	}

	signals := make(chan os.Signal)

	defer signal.Stop(signals)
	defer close(signals)

	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signals

		slog.Warn("received stop signal")

		err := a.Close()

		if err != nil {
			panic(err)
		}
	}()

	filteredArgs := a.filterArgs(args)

	err := command.Handle(context.Background(), filteredArgs)

	if err != nil {
		panic(err)
	}

	slog.Warn("Exit")
}

func (a *App) Close() error {
	slog.Warn("Closing app...")

	for _, listener := range append(a.closeListeners, a.lastCloseListeners...) {
		err := listener.Close()

		if err != nil {
			return errs.Err(err)
		}
	}

	return nil
}

func (a *App) AddFirstCloseListener(listener io.Closer) {
	a.closeListeners = append(a.closeListeners, listener)
}

func (a *App) AddLastCloseListener(listener io.Closer) {
	a.lastCloseListeners = append(a.closeListeners, listener)
}

func (a *App) initLogging() {
	customHandler, err := logging.NewCustomHandler(
		logging.NewLevelPolicy(a.config.LogConfig.Levels),
		a.config.LogConfig.DirPath,
		a.config.LogConfig.KeepDays,
	)

	if err == nil {
		slog.SetDefault(slog.New(customHandler))
	} else {
		panic(err)
	}

	a.AddLastCloseListener(customHandler)
}

func (a *App) filterArgs(args []string) []string {
	var result []string

	for _, arg := range args {
		if !strings.HasPrefix(arg, "--") {
			result = append(result, arg)
		}
	}

	return result
}
