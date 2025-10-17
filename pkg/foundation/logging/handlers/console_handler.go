package handlers

import (
	"context"
	"io"
	"log/slog"
	"os"
	"sparallel_server/pkg/foundation/errs"
)

var _ io.Closer = (*ConsoleHandler)(nil)

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

func (h *ConsoleHandler) Handle(_ context.Context, r slog.Record) error {
	msg := h.wrapColor(r.Level, makeMessageByRecord(r)) + "\n"

	_, err := os.Stdout.WriteString(msg)

	return errs.Err(err)
}

func (h *ConsoleHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (h *ConsoleHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *ConsoleHandler) WithGroup(_ string) slog.Handler {
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
