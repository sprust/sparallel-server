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
	"sync"
	"syscall"
	"time"
)

var _ io.Closer = (*App)(nil)

type App struct {
	config             *config.Config
	commands           map[string]commands.CommandInterface
	serviceProviders   []ServiceProviderInterface
	runningCommands    []commands.CommandInterface
	lastCloseListeners []io.Closer
	mu                 sync.RWMutex
}

func NewApp(
	config *config.Config,
	commands map[string]commands.CommandInterface,
	serviceProviders []ServiceProviderInterface,
) App {
	return App{
		config:           config,
		commands:         commands,
		serviceProviders: serviceProviders,
	}
}

func (a *App) Start(commandName string, args []string) {
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

	signals := make(chan os.Signal, 4)
	defer signal.Stop(signals)

	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGTSTP, syscall.SIGCONT)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)

	go func() {
		filteredArgs := a.filterArgs(args)

		done <- command.Handle(ctx, filteredArgs)
	}()

	for {
		select {
		case err := <-done:
			if err != nil {
				slog.Error("command failed", "error", err)
				panic(err)
			}
			slog.Info("command completed successfully")
			return

		case sgn := <-signals:
			switch sgn {
			case syscall.SIGTERM, os.Interrupt:
				if sgn == syscall.SIGTERM {
					slog.Warn("received stop (SIGTERM) signal")
				} else {
					slog.Warn("received interrupt signal (Ctrl+C)")
				}

				cancel()

				shutdownDone := make(chan error, 1)

				go func() {
					shutdownDone <- a.Close()
				}()

				select {
				case err := <-shutdownDone:
					if err != nil {
						slog.Error("error during shutdown", "error", err)
						os.Exit(1)
					}

					slog.Info("graceful shutdown completed")

					os.Exit(0)
				case <-time.After(10 * time.Second):
					slog.Error("shutdown timeout exceeded, forcing exit")

					os.Exit(1)
				}
			case syscall.SIGTSTP:
				slog.Warn("received pause (SIGTSTP) signal")

				if err := a.Pause(); err != nil {
					slog.Error("error pausing app", "error", err)
				}
			case syscall.SIGCONT:
				slog.Warn("received unpause (SIGCONT) signal")

				if err := a.UnPause(); err != nil {
					slog.Error("error unpausing app", "error", err)
				}
			}
		}
	}
}

func (a *App) Pause() error {
	slog.Warn("Pausing app...")

	a.mu.RLock()
	defer a.mu.RUnlock()

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

	a.mu.RLock()
	defer a.mu.RUnlock()

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

	a.mu.RLock()
	defer a.mu.RUnlock()

	var errsList []error

	for _, command := range a.runningCommands {
		if err := command.Close(); err != nil {
			slog.Error("error closing command", "error", err)
			errsList = append(errsList, err)
		}
	}

	for _, listener := range a.lastCloseListeners {
		if err := listener.Close(); err != nil {
			slog.Error("error closing listener", "error", err)
			errsList = append(errsList, err)
		}
	}

	if len(errsList) > 0 {
		return fmt.Errorf("close failed with %d errors: %v", len(errsList), errsList)
	}

	return nil
}

func (a *App) addRunningCommand(listener commands.CommandInterface) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.runningCommands = append(a.runningCommands, listener)
}

func (a *App) AddLastCloseListener(listener io.Closer) {
	a.mu.Lock()
	defer a.mu.Unlock()

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
