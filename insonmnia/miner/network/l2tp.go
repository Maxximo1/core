package network

import (
	"context"
	"fmt"
	"os/exec"
	"sync"

	"io/ioutil"

	"github.com/docker/go-plugins-helpers/network"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	XL2tpRunDir      = "/var/run/xl2tpd"
	XL2tpControlFile = "/var/run/xl2tpd/l2tp-control"
	PPPOptsDir       = "/etc/ppp/"
)

func NewL2TPDriver(ctx context.Context) *L2TPDriver {
	return &L2TPDriver{
		ctx:      ctx,
		networks: make(map[string]*networkInfo),
	}
}

type L2TPDriver struct {
	mu       sync.Mutex
	ctx      context.Context
	networks map[string]*networkInfo
}

func (d *L2TPDriver) GetCapabilities() (*network.CapabilitiesResponse, error) {
	log.G(d.ctx).Info("received  request", zap.Any("request", nil))
	return &network.CapabilitiesResponse{
		Scope:             "local",
		ConnectivityScope: "local",
	}, nil
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

	rawOpts, ok := request.Options["com.docker.network.generic"]
	if !ok {
		log.G(d.ctx).Error("no options provided")
		return errors.New("no options provided")
	}

	opts, ok := rawOpts.(map[string]interface{})
	if !ok {
		log.G(d.ctx).Error("invalid options")
		return errors.New("invalid options")
	}

	for optName, optVal := range opts {
		switch optName {
		case "lns_addr":
			if optValTyped, ok := optVal.(string); ok {
				netInfo.LNSAddr = optValTyped
			} else {
				log.G(d.ctx).Error("invalid option value", zap.String("option_name", optName),
					zap.Any("option_value", optVal))
				return fmt.Errorf("invalid value for option %s: %s", optName, optVal)
			}
		case "ppp_username":
			if optValTyped, ok := optVal.(string); ok {
				netInfo.PPPUsername = optValTyped
			} else {
				log.G(d.ctx).Error("invalid option value", zap.String("option_name", optName),
					zap.Any("option_value", optVal))
				return fmt.Errorf("invalid value for option %s: %s", optName, optVal)
			}
		case "ppp_password":
			if optValTyped, ok := optVal.(string); ok {
				netInfo.PPPPassword = optValTyped
			} else {
				log.G(d.ctx).Error("invalid option value", zap.String("option_name", optName),
					zap.Any("option_value", optVal))
				return fmt.Errorf("invalid value for option %s: %s", optName, optVal)
			}
		}
	}

	if err := netInfo.setup(); err != nil {
		return err
	}

	var (
		pppConfig = netInfo.getPPPConfig()
		xl2tpdCfg = netInfo.getXl2tpConfig()
		cmd       = exec.Command("xl2tpd-control", "add", netInfo.InternalName, xl2tpdCfg)
	)

	if err := ioutil.WriteFile(netInfo.PPPOptFile, []byte(pppConfig), 0644); err != nil {
		log.G(d.ctx).Error("failed to create PPP config file", zap.String("network_id", netInfo.ID),
			zap.Any("config", xl2tpdCfg), zap.Error(err))
		return errors.Wrapf(err, "failed to add xl2tpd config for network %s, config is `%s`",
			netInfo.ID, xl2tpdCfg)
	}

	if err := cmd.Run(); err != nil {
		log.G(d.ctx).Error("failed to add xl2tpd config", zap.String("network_id", netInfo.ID),
			zap.Any("config", xl2tpdCfg), zap.Error(err))
		return errors.Wrapf(err, "failed to add xl2tpd config for network %s, config is `%s`",
			netInfo.ID, xl2tpdCfg)
	}

	d.networks[netInfo.ID] = netInfo

	log.G(d.ctx).Info("successfully registered network", zap.String("network_id", netInfo.ID))

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
	return nil, nil
	//return &network.JoinResponse{DisableGatewayService: true, InterfaceName: network.InterfaceName{SrcName: "netname", DstPrefix: "pidor"}}, nil
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
	ID           string
	InternalName string
	IPv4Data     []*network.IPAMData
	IPv6Data     []*network.IPAMData

	LNSAddr     string
	PPPOptFile  string
	PPPUsername string
	PPPPassword string
}

func (i *networkInfo) setup() error {
	if len(i.LNSAddr) == 0 {
		return errors.New("lsn address not provided")
	}

	if len(i.PPPUsername) == 0 {
		return errors.New("ppp username not provided")
	}

	if len(i.PPPPassword) == 0 {
		return errors.New("ppp password not provided")
	}

	i.PPPOptFile = PPPOptsDir + i.ID[5:] + ".client"

	return nil
}

func (i *networkInfo) getPPPConfig() string {
	cfg := `ipcp-accept-local  
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

	cfg += fmt.Sprintf("\nname %s", i.PPPUsername)
	cfg += fmt.Sprintf("\npassword %s", i.PPPPassword)

	return cfg
}

func (i *networkInfo) getXl2tpConfig() string {
	return fmt.Sprintf("lns=%s pppoptfile=%s", i.LNSAddr, i.PPPOptFile)
}
