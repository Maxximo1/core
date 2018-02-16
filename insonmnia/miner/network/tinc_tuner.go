package network

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/sonm-io/core/insonmnia/miner/plugin"
)

type TincTuner struct {
	client *client.Client
}

func NewTincTuner(config *plugin.TincNetworkConfig) (*TincTuner, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	return &TincTuner{client: cli}, nil
}

func (t *TincTuner) Tune(ctx context.Context, net plugin.Network, config *network.NetworkingConfig) (plugin.Cleanup, error) {
	createOpts := types.NetworkCreate{
		Driver:  "tinc",
		Options: net.NetworkOptions(),
	}
	if len(net.NetworkCIDR()) != 0 {
		createOpts.IPAM = &network.IPAM{
			Driver: "default",
			Config: make([]network.IPAMConfig, 0),
		}
		createOpts.IPAM.Config = append(createOpts.IPAM.Config, network.IPAMConfig{Subnet: net.NetworkCIDR()})
	}
	response, err := t.client.NetworkCreate(ctx, net.ID(), createOpts)
	if err != nil {
		return err
	}
	if config.EndpointsConfig == nil {
		config.EndpointsConfig = make(map[string]*network.EndpointSettings)
		config.EndpointsConfig[response.ID] = &network.EndpointSettings{
			IPAddress: net.NetworkAddr(),
		}
	}

	return nil
}
