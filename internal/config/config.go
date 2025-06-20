package config

import (
	fConfig "sparallel_server/pkg/foundation/config"
)

type Config struct {
	fCfg *fConfig.Config
}

var config *Config

func Init(fCfg *fConfig.Config) *Config {
	if config != nil {
		panic("Config is already initialized")
	}

	config = &Config{
		fCfg: fCfg,
	}

	return config
}

func GetConfig() *Config {
	if config == nil {
		panic("Config is not initialized")
	}

	return config
}

func (c *Config) Reload() *Config {
	c.fCfg.Load()

	return c
}

func (c *Config) GetServerPidFilePath() string {
	return c.fCfg.GetString("SERVER_PID_FILE_PATH")
}

func (c *Config) GetRpcPort() string {
	return c.fCfg.GetString("RPC_PORT")
}

func (c *Config) GetCommand() string {
	return c.fCfg.GetString("WORKER_COMMAND")
}

func (c *Config) IsServeProxy() bool {
	return c.fCfg.GetBool("SERVE_PROXY")
}

func (c *Config) IsServeWorkers() bool {
	return c.fCfg.GetBool("SERVE_WORKERS")
}

func (c *Config) GetMinWorkersNumber() int {
	return c.fCfg.GetInt("MIN_WORKERS_NUMBER")
}

func (c *Config) GetMaxWorkersNumber() int {
	return c.fCfg.GetInt("MAX_WORKERS_NUMBER")
}

func (c *Config) GetWorkersNumberScaleUp() int {
	return c.fCfg.GetInt("WORKERS_NUMBER_SCALE_UP")
}

func (c *Config) GetWorkersNumberPercentScaleUp() int {
	return c.fCfg.GetInt("WORKERS_NUMBER_PERCENT_SCALE_UP")
}

func (c *Config) GetWorkersNumberPercentScaleDown() int {
	return c.fCfg.GetInt("WORKERS_NUMBER_PERCENT_SCALE_DOWN")
}
