package network

import (
	"context"
	"errors"
	"sync"

	"fmt"

	"github.com/docker/go-plugins-helpers/network"
	log "github.com/noxiouz/zapctx/ctxlog"
	"go.uber.org/zap"
)

const (
	XL2tpRunDir      = "/var/run/xl2tpd"
	XL2tpControlFile = "/var/run/xl2tpd/l2tp-control"
	PPPOptsDir       = "/etc/ppp/"
	PPPOptsDefaukt   = `ipcp-accept-local  
ipcp-accept-remote  
refuse-eap  
require-mschap-v2  
noccp  
noauth  
idle 1800  
mtu 1410  
mru 1410  
defaultroute  
usepeerdns  
debug  
lock  
connect-delay 5000`
)

func NewL2TPDriver(ctx context.Context) *L2TPDriver {
	return &L2TPDriver{ctx: ctx}
}

type L2TPDriver struct {
	mu       sync.Mutex
	ctx      context.Context
	networks map[string]*networkInfo
}

func (d *L2TPDriver) GetCapabilities() (*network.CapabilitiesResponse, error) {
	return &network.CapabilitiesResponse{
		Scope:             "global",
		ConnectivityScope: "global",
	}, nil
	log.G(d.ctx).Info("received  request", zap.Any("request", nil))
	return nil, nil
}

func (d *L2TPDriver) CreateNetwork(request *network.CreateNetworkRequest) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	log.G(d.ctx).Info("received CreateNetwork request", zap.Any("request", request))

	netInfo := &networkInfo{
		ID:       request.NetworkID,
		IPv4Data: request.IPv4Data,
		IPv6Data: request.IPv6Data,
	}

	if _, ok := d.networks[netInfo.ID]; ok {
		return errors.New("network already exists")
	}

	for optName, optVal := range request.Options {
		switch optName {
		case "lns_addr":
			if optValTyped, ok := optVal.(string); ok {
				netInfo.LNSAddr = optValTyped
			} else {
				log.G(d.ctx).Warn("invalid option value", zap.String("option_name", optName),
					zap.Any("option_value", optVal))
				return fmt.Errorf("invalid value for option %s: %s", optName, optVal)
			}
		case "ppp_username":
			if optValTyped, ok := optVal.(string); ok {
				netInfo.PPPUsername = optValTyped
			} else {
				log.G(d.ctx).Warn("invalid option value", zap.String("option_name", optName),
					zap.Any("option_value", optVal))
				return fmt.Errorf("invalid value for option %s: %s", optName, optVal)
			}
		case "ppp_password":
			if optValTyped, ok := optVal.(string); ok {
				netInfo.PPPPassword = optValTyped
			} else {
				log.G(d.ctx).Warn("invalid option value", zap.String("option_name", optName),
					zap.Any("option_value", optVal))
				return fmt.Errorf("invalid value for option %s: %s", optName, optVal)
			}
		}
	}

	if err := netInfo.setup(); err != nil {
		return err
	}

	d.networks[netInfo.ID] = netInfo

	return nil
}

func (d *L2TPDriver) AllocateNetwork(request *network.AllocateNetworkRequest) (*network.AllocateNetworkResponse, error) {
	log.G(d.ctx).Info("received AllocateNetwork request", zap.Any("request", request))
	return nil, nil
}

func (d *L2TPDriver) DeleteNetwork(request *network.DeleteNetworkRequest) error {
	log.G(d.ctx).Info("received DeleteNetwork request", zap.Any("request", request))
	return nil
}

func (d *L2TPDriver) FreeNetwork(request *network.FreeNetworkRequest) error {
	log.G(d.ctx).Info("received FreeNetwork request", zap.Any("request", request))
	return nil
}

func (d *L2TPDriver) CreateEndpoint(request *network.CreateEndpointRequest) (*network.CreateEndpointResponse, error) {
	log.G(d.ctx).Info("received CreateEndpoint request", zap.Any("request", request))
	return &network.CreateEndpointResponse{}, nil
	//return &network.CreateEndpointResponse{Interface: request.Interface}, nil
}

func (d *L2TPDriver) DeleteEndpoint(request *network.DeleteEndpointRequest) error {
	log.G(d.ctx).Info("received DeleteEndpoint request", zap.Any("request", request))
	return nil
}

func (d *L2TPDriver) EndpointInfo(request *network.InfoRequest) (*network.InfoResponse, error) {
	log.G(d.ctx).Info("received EndpointInfo request", zap.Any("request", request))
	val := make(map[string]string)
	val["Endpoint"] = "10.274.0.4/16"
	return &network.InfoResponse{Value: val}, nil
}

func (d *L2TPDriver) Join(request *network.JoinRequest) (*network.JoinResponse, error) {
	log.G(d.ctx).Info("received Join request", zap.Any("request", request))
	//return nil, nil
	return &network.JoinResponse{DisableGatewayService: true, InterfaceName: network.InterfaceName{SrcName: "netname", DstPrefix: "pidor"}}, nil
}

func (d *L2TPDriver) Leave(request *network.LeaveRequest) error {
	log.G(d.ctx).Info("received Leave request", zap.Any("request", request))
	return nil
}

func (d *L2TPDriver) DiscoverNew(request *network.DiscoveryNotification) error {
	log.G(d.ctx).Info("received DiscoverNew request", zap.Any("request", request))
	return nil
}

func (d *L2TPDriver) DiscoverDelete(request *network.DiscoveryNotification) error {
	log.G(d.ctx).Info("received DiscoverDelete request", zap.Any("request", request))
	return nil
}

func (d *L2TPDriver) ProgramExternalConnectivity(request *network.ProgramExternalConnectivityRequest) error {
	log.G(d.ctx).Info("received ProgramExternalConnectivity request", zap.Any("request", request))
	return nil
}

func (d *L2TPDriver) RevokeExternalConnectivity(request *network.RevokeExternalConnectivityRequest) error {
	log.G(d.ctx).Info("received RevokeExternalConnectivity request", zap.Any("request", request))
	return nil
}

type networkInfo struct {
	ID       string
	IPv4Data []*network.IPAMData
	IPv6Data []*network.IPAMData

	LNSAddr     string
	PPPOptFile  string
	PPPUsername string
	PPPPassword string
}

func (i *networkInfo) setup() error {
	i.PPPOptFile = PPPOptsDir + i.ID[5:]
	return nil
}
