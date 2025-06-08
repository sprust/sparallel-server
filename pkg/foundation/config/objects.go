package config

import (
	"github.com/joho/godotenv"
	"log/slog"
	"sparallel_server/pkg/foundation/errs"
	"sync"
)

var config *Config
var once sync.Once

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

func Init(filenames ...string) {
	once.Do(func() {
		config = &Config{
			filenames: filenames,
		}

		config.Load()
	})
}

func GetConfig() *Config {
	if config == nil {
		panic("Config is not initialized")
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
}
