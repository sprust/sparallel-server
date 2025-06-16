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
	config             *config.Config
	commands           map[string]commands.CommandInterface
	serviceProviders   []ServiceProviderInterface
	runningCommands    []commands.CommandInterface
	lastCloseListeners []io.Closer
}

func NewApp(
	config *config.Config,
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

			slog.Error(fmt.Sprintf("%s", r))

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

	a.addRunningCommand(command)

	for _, provider := range a.serviceProviders {
		err := provider.Register()

		if err != nil {
			panic(err)
		}
	}

	signals := make(chan os.Signal, 3)

	defer signal.Stop(signals)
	defer close(signals)

	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	signal.Notify(signals, os.Interrupt, syscall.SIGTSTP)
	signal.Notify(signals, os.Interrupt, syscall.SIGCONT)

	go func() {
		for {
			sgn := <-signals

			var err error

			switch sgn {
			case syscall.SIGTERM:
				slog.Warn("received stop (SIGTERM) signal")

				_ = a.Close()

				return
			case syscall.SIGTSTP:
				slog.Warn("received pause (SIGTSTP) signal")

				err = a.Pause()
			case syscall.SIGCONT:
				slog.Warn("received unpause (SIGCONT) signal")

				err = a.UnPause()
			default:
				if sgn == nil {
					slog.Warn("received nil signal")
				} else {
					slog.Warn("received unknown signal: " + sgn.String())
				}
			}

			if err != nil {
				panic(err)
			}
		}
	}()

	filteredArgs := a.filterArgs(args)

	err := command.Handle(context.Background(), filteredArgs)

	if err != nil {
		panic(err)
	}

	slog.Warn("Exit")
}

func (a *App) Pause() error {
	slog.Warn("Pausing app...")

	for _, listener := range a.runningCommands {
		err := listener.Pause()

		if err != nil {
			return errs.Err(err)
		}
	}

	return nil
}

func (a *App) UnPause() error {
	slog.Warn("Unpausing app...")

	for _, command := range a.runningCommands {
		err := command.UnPause()

		if err != nil {
			return errs.Err(err)
		}
	}

	return nil
}

func (a *App) Close() error {
	slog.Warn("Closing app...")

	for _, command := range a.runningCommands {
		err := command.Close()

		if err != nil {
			return errs.Err(err)
		}
	}

	for _, listener := range a.lastCloseListeners {
		err := listener.Close()

		if err != nil {
			return errs.Err(err)
		}
	}

	return nil
}

func (a *App) addRunningCommand(listener commands.CommandInterface) {
	a.runningCommands = append(a.runningCommands, listener)
}

func (a *App) AddLastCloseListener(listener io.Closer) {
	a.lastCloseListeners = append(a.lastCloseListeners, listener)
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
