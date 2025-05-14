package logging

import "log/slog"

type LevelPolicy struct {
	any   bool
	debug bool
	info  bool
	warn  bool
	error bool
}

func NewLevelPolicy(levels []slog.Level) *LevelPolicy {
	p := &LevelPolicy{}

	if len(levels) == 0 {
		p.any = true
		return p
	}

	for _, level := range levels {
		switch level {
		case slog.LevelDebug:
			p.debug = true
		case slog.LevelInfo:
			p.info = true
		case slog.LevelWarn:
			p.warn = true
		case slog.LevelError:
			p.error = true
		}
	}

	return p
}

func (l *LevelPolicy) Allowed(level slog.Level) bool {
	if l.any {
		return true
	}

	switch level {
	case slog.LevelDebug:
		return l.debug
	case slog.LevelInfo:
		return l.info
	case slog.LevelWarn:
		return l.warn
	case slog.LevelError:
		return l.error
	default:
		return false
	}
}
