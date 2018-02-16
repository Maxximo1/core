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
	driver *TincNetworkDriver
}

type TincCleaner struct {
	networkID string
	client    *client.Client
}

func NewTincTuner(ctx context.Context, config *plugin.TincNetworkConfig) (*TincTuner, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	driver, err := NewTincNetworkDriver(ctx, config)
	if err != nil {
		return nil, err
	}
	return &TincTuner{
		client: cli,
		driver: driver,
	}, nil
}

//TODO: pass context from outside
func (t *TincTuner) Tune(net plugin.Network, config *network.NetworkingConfig) (plugin.Cleanup, error) {
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
	response, err := t.client.NetworkCreate(context.Background(), net.ID(), createOpts)
	if err != nil {
		return nil, err
	}
	if config.EndpointsConfig == nil {
		config.EndpointsConfig = make(map[string]*network.EndpointSettings)
		config.EndpointsConfig[response.ID] = &network.EndpointSettings{
			IPAddress: net.NetworkAddr(),
		}
	}

	return &TincCleaner{
		client:    t.client,
		networkID: response.ID,
	}, nil
}

func (t *TincCleaner) Close() error {
	return t.client.NetworkRemove(context.Background(), t.networkID)
}
