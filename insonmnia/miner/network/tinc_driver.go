package network

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/docker/go-plugins-helpers/network"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/insonmnia/miner/plugin"
	"go.uber.org/zap"
)

func NewTincNetworkDriver(ctx context.Context, config *plugin.TincNetworkConfig) (*TincNetworkDriver, error) {
	err := os.MkdirAll(config.ConfigDir, 0770)
	if err != nil {
		return nil, err
	}

	return &TincNetworkDriver{
		ctx:      ctx,
		config:   config,
		networks: make(map[string]*TincNetwork),
	}, nil
}

type TincNetworkOptions struct {
	Invitation   string
	EnableBridge bool
}

type TincNetwork struct {
	ID         string
	Options    *TincNetworkOptions
	IPv4Data   []*network.IPAMData
	IPv6Data   []*network.IPAMData
	ConfigPath string
}

type TincNetworkDriver struct {
	ctx      context.Context
	config   *plugin.TincNetworkConfig
	mu       sync.RWMutex
	networks map[string]*TincNetwork
}

func (t *TincNetworkDriver) newTincNetwork(request *network.CreateNetworkRequest) (*TincNetwork, error) {
	opts, err := ParseNetworkOpts(request.Options)
	if err != nil {
		return nil, err
	}
	configPath := t.config.ConfigDir + "/" + request.NetworkID
	err = os.MkdirAll(configPath, 0770)
	if err != nil {
		return nil, err
	}

	return &TincNetwork{
		ID:       request.NetworkID,
		Options:  opts,
		IPv4Data: request.IPv4Data,
		IPv6Data: request.IPv6Data,
		// TODO: configurable
		ConfigPath: configPath,
	}, nil
}

func ParseNetworkOpts(data map[string]interface{}) (*TincNetworkOptions, error) {
	g, ok := data["com.docker.network.generic"]
	if !ok {
		return nil, errors.New("no options passed - invitation is required")
	}
	generic, ok := g.(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid type of generic options")
	}
	invitation, ok := generic["invitation"]
	if !ok {
		return nil, errors.New("invitation is required")
	}
	_, ok = generic["enable_bridge"]
	return &TincNetworkOptions{Invitation: invitation.(string), EnableBridge: ok}, nil
}

func (t *TincNetworkDriver) GetCapabilities() (*network.CapabilitiesResponse, error) {
	return &network.CapabilitiesResponse{
		Scope:             "global",
		ConnectivityScope: "global",
	}, nil
	log.G(t.ctx).Info("received  request", zap.Any("request", nil))
	return nil, nil
}

func (t *TincNetworkDriver) CreateNetwork(request *network.CreateNetworkRequest) error {
	log.G(t.ctx).Info("received CreateNetwork request", zap.Any("request", request))
	network, err := t.newTincNetwork(request)
	if err != nil {
		return err
	}

	time.Sleep(time.Second)

	err = runCommand(t.ctx, "tinc", "--batch", "-n", network.ID, "-c", network.ConfigPath, "join", network.Options.Invitation)
	if err != nil {
		return err
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	t.networks[network.ID] = network

	return nil
}

func (t *TincNetworkDriver) AllocateNetwork(request *network.AllocateNetworkRequest) (*network.AllocateNetworkResponse, error) {
	log.G(t.ctx).Info("received AllocateNetwork request", zap.Any("request", request))
	return nil, nil
}

func (t *TincNetworkDriver) popNetwork(ID string) *TincNetwork {
	t.mu.Lock()
	defer t.mu.Unlock()
	net, ok := t.networks[ID]
	if !ok {
		return nil
	}
	delete(t.networks, ID)
	return net
}

func (t *TincNetworkDriver) shutdownNetwork(network *TincNetwork) error {
	err := os.RemoveAll(network.ConfigPath)
	if err != nil {
		return err
	}
	return nil
}

func (t *TincNetworkDriver) DeleteNetwork(request *network.DeleteNetworkRequest) error {
	net := t.popNetwork(request.NetworkID)
	if net == nil {
		return errors.Errorf("no network with id %s", request.NetworkID)
	}
	t.shutdownNetwork(net)
	return nil
}

func (t *TincNetworkDriver) FreeNetwork(request *network.FreeNetworkRequest) error {
	log.G(t.ctx).Info("received FreeNetwork request", zap.Any("request", request))
	return nil
}

func (t *TincNetworkDriver) CreateEndpoint(request *network.CreateEndpointRequest) (*network.CreateEndpointResponse, error) {
	log.G(t.ctx).Info("received CreateEndpoint request", zap.Any("request", request))

	t.mu.RLock()
	defer t.mu.RUnlock()
	net, ok := t.networks[request.NetworkID]
	if !ok {
		return nil, errors.Errorf("no such network %s", request.NetworkID)
	}
	iface := net.ID[:15]
	//TODO: each pool should be considered
	pool := net.IPv4Data[0].Pool
	selfAddr := strings.Split(request.Interface.Address, "/")[0]
	err := runCommand(t.ctx, "tinc", "-n", net.ID, "-c", net.ConfigPath, "start",
		"-o", "Interface="+iface, "-o", "Subnet="+pool, "-o", "Subnet="+selfAddr+"/32")
	if err != nil {
		return nil, err
	}

	return &network.CreateEndpointResponse{}, nil
}

func (t *TincNetworkDriver) DeleteEndpoint(request *network.DeleteEndpointRequest) error {
	log.G(t.ctx).Info("received DeleteEndpoint request", zap.Any("request", request))

	t.mu.RLock()
	defer t.mu.RUnlock()

	net, ok := t.networks[request.NetworkID]
	if !ok {
		return errors.Errorf("no such network %s", request.NetworkID)
	}
	err := runCommand(t.ctx, "tinc", "--batch", "-n", net.ID, "-c", net.ConfigPath, "stop")
	if err != nil {
		return err
	}

	return nil
}

func (t *TincNetworkDriver) EndpointInfo(request *network.InfoRequest) (*network.InfoResponse, error) {
	log.G(t.ctx).Info("received EndpointInfo request", zap.Any("request", request))
	val := make(map[string]string)
	return &network.InfoResponse{Value: val}, nil
}

func (t *TincNetworkDriver) Join(request *network.JoinRequest) (*network.JoinResponse, error) {
	log.G(t.ctx).Info("received Join request", zap.Any("request", request))
	t.mu.RLock()
	defer t.mu.RUnlock()
	iface := request.NetworkID[:15]
	return &network.JoinResponse{DisableGatewayService: true, InterfaceName: network.InterfaceName{SrcName: iface, DstPrefix: "tinc"}}, nil
}

func (t *TincNetworkDriver) Leave(request *network.LeaveRequest) error {
	log.G(t.ctx).Info("received Leave request", zap.Any("request", request))
	return nil
}

func (t *TincNetworkDriver) DiscoverNew(request *network.DiscoveryNotification) error {
	log.G(t.ctx).Info("received DiscoverNew request", zap.Any("request", request))
	return nil
}

func (t *TincNetworkDriver) DiscoverDelete(request *network.DiscoveryNotification) error {
	log.G(t.ctx).Info("received DiscoverDelete request", zap.Any("request", request))
	return nil
}

func (t *TincNetworkDriver) ProgramExternalConnectivity(request *network.ProgramExternalConnectivityRequest) error {
	log.G(t.ctx).Info("received ProgramExternalConnectivity request", zap.Any("request", request))
	return nil
}

func (t *TincNetworkDriver) RevokeExternalConnectivity(request *network.RevokeExternalConnectivityRequest) error {
	log.G(t.ctx).Info("received RevokeExternalConnectivity request", zap.Any("request", request))
	return nil
}

func runCommand(ctx context.Context, name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.S(ctx).Errorf("failed to execute command - %s %s, output - %s", name, arg, out)
		return err
	} else {
		log.S(ctx).Info("executed command", zap.String("cmd", name), zap.Any("args", arg), zap.String("output", (string)(out)))
	}
	return nil
}
