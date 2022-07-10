package miner

import (
	"runtime"

	"github.com/go-logr/logr"
	"github.com/stratumfarm/dero-stratum-miner/internal/version"
)

func (c *Client) setLogger(logger logr.Logger) error {
	c.logger = logger.WithName("miner")
	c.logger.Info("DERO Stargate HE AstroBWT stratum miner")
	c.logger.Info("", "OS", runtime.GOOS, "ARCH", runtime.GOARCH, "GOMAXPROCS", runtime.GOMAXPROCS(0))
	c.logger.Info("", "Version", version.Version)

	name := "mainnet"
	if c.config.Testnet {
		name = "testnet"
	}
	c.logger.V(0).Info("", "MODE", name)
	return nil
}
