package config

import (
	"os"
	"strconv"
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

func (c *Config) GetCommand() string {
	return os.Getenv("WORKER_COMMAND")
}

func (c *Config) GetMinWorkersNumber() int {
	value, _ := strconv.Atoi(os.Getenv("MIN_WORKERS_NUMBER"))
	return value
}

func (c *Config) GetMaxWorkersNumber() int {
	value, _ := strconv.Atoi(os.Getenv("MAX_WORKERS_NUMBER"))
	return value
}

func (c *Config) GetWorkersNumberPercentScale() int {
	value, _ := strconv.Atoi(os.Getenv("WORKERS_NUMBER_PERCENT_SCALE"))
	return value
}

func (c *Config) GetWorkersNumberScaleUp() int {
	value, _ := strconv.Atoi(os.Getenv("WORKERS_NUMBER_SCALE_UP"))
	return value
}

func (c *Config) GetWorkersNumberScaleDown() int {
	value, _ := strconv.Atoi(os.Getenv("WORKERS_NUMBER_SCALE_DOWN"))
	return value
}
