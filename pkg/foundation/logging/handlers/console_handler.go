package handlers

import (
	"context"
	"log/slog"
	"os"
	"sparallel_server/pkg/foundation/errs"
)

const (
	reset  = "\033[0m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
)

type ConsoleHandler struct {
}

func NewConsoleHandler() *ConsoleHandler {
	return &ConsoleHandler{}
}

func (h *ConsoleHandler) Handle(ctx context.Context, r slog.Record) error {
	msg := h.wrapColor(r.Level, makeMessageByRecord(r)) + "\n"

	_, err := os.Stdout.WriteString(msg)

	return errs.Err(err)
}

func (h *ConsoleHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (h *ConsoleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *ConsoleHandler) WithGroup(name string) slog.Handler {
	return h
}

func (h *ConsoleHandler) Close() error {
	return nil
}

func (h *ConsoleHandler) wrapColor(level slog.Level, msg string) string {
	var wrapText string

	switch level {
	case slog.LevelDebug:
		wrapText = blue
	case slog.LevelInfo:
		wrapText = green
	case slog.LevelWarn:
		wrapText = yellow
	case slog.LevelError:
		wrapText = red
	default:
		return msg
	}

	return wrapText + msg + reset
}
