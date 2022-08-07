package miner

import (
	"fmt"
	"runtime"

	"github.com/go-logr/logr"
	"github.com/whalesburg/dero-stratum-miner/internal/version"
)

func (c *Client) setLogger(logger logr.Logger) {
	c.logger = logger.WithName("miner")
	c.logger.Info("AstroBWT Stratum Miner")
	c.logger.Info("Version: " + version.Version)
	c.logger.Info("OS: " + runtime.GOOS)
	c.logger.Info("Arch: " + runtime.GOARCH)
	c.logger.Info(fmt.Sprintf("Threads: %d (max: %d)", c.config.Threads, runtime.GOMAXPROCS(0)))

	name := "mainnet"
	if c.config.Testnet {
		name = "testnet"
	}
	c.logger.Info("Network: " + name)
}
