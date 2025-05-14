package logging

import (
	"context"
	"log/slog"
	"sparallel_server/pkg/foundation/errs"
	"sparallel_server/pkg/foundation/logging/handlers"
)

type CustomHandler struct {
	levelPolicy    *LevelPolicy
	consoleHandler *handlers.ConsoleHandler
	fileHandler    *handlers.FileHandler
}

func NewCustomHandler(levelPolicy *LevelPolicy, logDirPath string, logKeepDays int) (*CustomHandler, error) {
	fileHandler, err := handlers.NewFileHandler(logDirPath, logKeepDays)

	if err != nil {
		return nil, err
	}

	handler := &CustomHandler{
		levelPolicy:    levelPolicy,
		fileHandler:    fileHandler,
		consoleHandler: handlers.NewConsoleHandler(),
	}

	return handler, nil
}

func (h *CustomHandler) Handle(ctx context.Context, r slog.Record) error {
	if err := h.consoleHandler.Handle(ctx, r); err != nil {
		return errs.Err(err)
	}
	if err := h.fileHandler.Handle(ctx, r); err != nil {
		return errs.Err(err)
	}

	return nil
}

func (h *CustomHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.levelPolicy.Allowed(level)
}

func (h *CustomHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *CustomHandler) WithGroup(name string) slog.Handler {
	return h
}

func (h *CustomHandler) Close() error {
	if err := h.fileHandler.Close(); err != nil {
		return errs.Err(err)
	}

	return nil
}
