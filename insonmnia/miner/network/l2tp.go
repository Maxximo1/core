// +build linux

package network

import (
	"context"
	"fmt"

	"crypto/md5"
	"encoding/hex"

	"github.com/docker/go-plugins-helpers/network"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	pppOptsDir = "/etc/ppp/"
)

func NewL2TPDriver(ctx context.Context) *L2TPDriver {
	return &L2TPDriver{
		ctx:      ctx,
		networks: make(map[string]*networkInfo),
	}
}

type L2TPDriver struct {
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
	log.G(d.ctx).Info("received CreateNetwork request", zap.Any("request", request))

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

	var (
		lnsAddr     string
		pppUsername string
		pppPassword string
	)
	for optName, optVal := range opts {
		switch optName {
		case "lns_addr":
			if optValTyped, ok := optVal.(string); ok {
				lnsAddr = optValTyped
			} else {
				log.G(d.ctx).Error("invalid option value", zap.String("option_name", optName),
					zap.Any("option_value", optVal))
				return fmt.Errorf("invalid value for option %s: %s", optName, optVal)
			}
		case "ppp_username":
			if optValTyped, ok := optVal.(string); ok {
				pppUsername = optValTyped
			} else {
				log.G(d.ctx).Error("invalid option value", zap.String("option_name", optName),
					zap.Any("option_value", optVal))
				return fmt.Errorf("invalid value for option %s: %s", optName, optVal)
			}
		case "ppp_password":
			if optValTyped, ok := optVal.(string); ok {
				pppPassword = optValTyped
			} else {
				log.G(d.ctx).Error("invalid option value", zap.String("option_name", optName),
					zap.Any("option_value", optVal))
				return fmt.Errorf("invalid value for option %s: %s", optName, optVal)
			}
		}
	}

	netInfo, ok := d.networks[GetMD5Hash(lnsAddr+pppUsername+pppPassword)]
	if !ok {
		return errors.New("unexpected network parameters")
	}

	netInfo.id = request.NetworkID
	d.networks[netInfo.id] = netInfo

	log.G(d.ctx).Info("successfully registered network", zap.String("network_id", netInfo.id))

	return nil
}

func (d *L2TPDriver) CreateEndpoint(request *network.CreateEndpointRequest) (*network.CreateEndpointResponse, error) {
	log.G(d.ctx).Info("received CreateEndpoint request", zap.Any("request", request))

	return nil, nil
}

func (d *L2TPDriver) Join(request *network.JoinRequest) (*network.JoinResponse, error) {
	log.G(d.ctx).Info("received Join request", zap.Any("request", request))

	netInfo, ok := d.networks[request.NetworkID]
	if !ok {
		return nil, errors.Errorf("network %s not found", request.NetworkID)
	}

	return &network.JoinResponse{
		InterfaceName: network.InterfaceName{SrcName: netInfo.endpoint.pppDevName, DstPrefix: "wut"},
	}, nil
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
	id         string
	internalID string
	ipv4Data   []*network.IPAMData
	ipv6Data   []*network.IPAMData

	lnsAddr     string
	pppUsername string
	pppPassword string

	endpoint *endpointInfo
}

func (i *networkInfo) setup() error {
	if len(i.lnsAddr) == 0 {
		return errors.New("lsn address not provided")
	}

	if len(i.pppUsername) == 0 {
		return errors.New("ppp username not provided")
	}

	if len(i.pppPassword) == 0 {
		return errors.New("ppp password not provided")
	}

	return nil
}

type endpointInfo struct {
	id                string
	internalID        string
	networkID         string
	networkInternalID string
	lnsAddr           string
	connName          string
	pppDevName        string
	pppOptFile        string
	pppUsername       string
	pppPassword       string
	assignedIP        string
}

func (i *endpointInfo) setup() error {
	i.connName = i.internalID + "-connection"
	i.pppDevName = ("ppp" + i.internalID)[:15]
	i.pppOptFile = pppOptsDir + i.internalID + "." + ".client"

	return nil
}

func (i *endpointInfo) getPppConfig() string {
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

	cfg += fmt.Sprintf("\nifname %s", i.pppDevName)
	cfg += fmt.Sprintf("\nname %s", i.pppUsername)
	cfg += fmt.Sprintf("\npassword %s", i.pppPassword)

	return cfg
}

func (i *endpointInfo) getXl2tpConfig() []string {
	return []string{
		fmt.Sprintf("lns=%s", i.lnsAddr),
		fmt.Sprintf("pppoptfile=%s", i.pppOptFile),
	}
}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
