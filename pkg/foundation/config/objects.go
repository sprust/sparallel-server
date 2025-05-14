package config

import "log/slog"

type Config struct {
	LogConfig LogConfig
}

type LogConfig struct {
	Levels   []slog.Level
	DirPath  string
	KeepDays int
}
