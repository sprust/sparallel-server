package config

import (
	"os"
	"sync"
)

type Config struct {
	mutex sync.Mutex
}

var instance *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		instance = &Config{}
	})

	return instance
}

func (c *Config) GetLogLevels() string {
	return os.Getenv("LOG_LEVELS")
}

func (c *Config) GetRpcPort() string {
	return os.Getenv("RPC_PORT")
}
