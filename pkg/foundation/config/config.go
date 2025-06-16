package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"log/slog"
	"os"
	"slices"
	"sparallel_server/pkg/foundation/errs"
	"strconv"
	"strings"
	"sync"
)

var config *Config

type Config struct {
	mutex     sync.Mutex
	filenames []string
	LogConfig LogConfig
}

type LogConfig struct {
	Levels   []slog.Level
	DirPath  string
	KeepDays int
}

func Init(filenames ...string) *Config {
	if config != nil {
		panic("App config is already initialized")
	}

	config = &Config{
		filenames: filenames,
	}

	config.Load()

	return config
}

func GetConfig() *Config {
	if config == nil {
		panic("Foundation config is not initialized")
	}

	return config
}

func (c *Config) Load() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	err := godotenv.Overload(c.filenames...)

	if err != nil {
		panic(errs.Err(err))
	}

	config.initLogLevels()
}

func (c *Config) GetString(key string) string {
	return os.Getenv(key)
}

func (c *Config) GetInt(key string) int {
	value, err := strconv.Atoi(os.Getenv(key))

	if err != nil {
		panic(errs.Err(err))
	}

	return value
}

func (c *Config) GetBool(key string) bool {
	return os.Getenv(key) == "true"
}

func (c *Config) initLogLevels() {
	logLevels := strings.Split(os.Getenv("LOG_LEVELS"), ",")

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

	logKeepDays, err := strconv.Atoi(os.Getenv("LOG_KEEP_DAYS"))

	if err != nil {
		logKeepDays = 3

		slog.Warn("LOG_KEEP_DAYS is not set or invalid, using default value [3]")
	}

	c.LogConfig = LogConfig{
		Levels:   slogLevels,
		DirPath:  os.Getenv("LOG_DIR"),
		KeepDays: logKeepDays,
	}
}
