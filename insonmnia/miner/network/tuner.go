package network

import (
	"github.com/docker/docker/api/types/network"
	"github.com/sonm-io/core/insonmnia/miner/plugin"
)

// Tuner is responsible for preparing GPU-friendly environment and baking proper options in container.HostConfig
type Tuner interface {
	Tune(net plugin.Network, config *network.NetworkingConfig) (plugin.Cleanup, error)
}
