// +build linux

package network

import (
	"context"
	"fmt"
	"io/ioutil"
	"os/exec"

	"time"

	"github.com/docker/go-plugins-helpers/network"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"
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

	netInfo := &networkInfo{
		id:        request.NetworkID,
		ipv4Data:  request.IPv4Data,
		ipv6Data:  request.IPv6Data,
		endpoints: make(map[string]*endpointInfo),
	}

	if _, ok := d.networks[netInfo.id]; ok {
		log.G(d.ctx).Error("network already exists", zap.String("network_id", request.NetworkID))
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
				netInfo.lnsAddr = optValTyped
			} else {
				log.G(d.ctx).Error("invalid option value", zap.String("option_name", optName),
					zap.Any("option_value", optVal))
				return fmt.Errorf("invalid value for option %s: %s", optName, optVal)
			}
		case "ppp_username":
			if optValTyped, ok := optVal.(string); ok {
				netInfo.pppUsername = optValTyped
			} else {
				log.G(d.ctx).Error("invalid option value", zap.String("option_name", optName),
					zap.Any("option_value", optVal))
				return fmt.Errorf("invalid value for option %s: %s", optName, optVal)
			}
		case "ppp_password":
			if optValTyped, ok := optVal.(string); ok {
				netInfo.pppPassword = optValTyped
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

	d.networks[netInfo.id] = netInfo

	log.G(d.ctx).Info("successfully registered network", zap.String("network_id", netInfo.id))

	return nil
}

func (d *L2TPDriver) CreateEndpoint(request *network.CreateEndpointRequest) (*network.CreateEndpointResponse, error) {
	log.G(d.ctx).Info("received CreateEndpoint request", zap.Any("request", request))

	netInfo, ok := d.networks[request.NetworkID]
	if !ok {
		log.G(d.ctx).Error("network not found", zap.String("network_id", request.NetworkID))
		return nil, errors.Errorf("network not found: %s", request.NetworkID)
	}

	if _, ok := netInfo.endpoints[request.EndpointID]; ok {
		log.G(d.ctx).Error("endpoint already exists", zap.String("network_id", request.NetworkID))
		return nil, errors.New("endpoint already exists")
	}

	eptInfo := &endpointInfo{
		id:          request.EndpointID,
		networkID:   netInfo.id,
		lnsAddr:     netInfo.lnsAddr,
		pppUsername: netInfo.pppUsername,
		pppPassword: netInfo.pppPassword,
	}

	if err := eptInfo.setup(); err != nil {
		return nil, err
	}

	var (
		pppCfg       = eptInfo.getPppConfig()
		xl2tpdCfg    = eptInfo.getXl2tpConfig()
		addCfgCmd    = exec.Command("xl2tpd-control", "add", eptInfo.сonnName, xl2tpdCfg[0], xl2tpdCfg[1])
		setupConnCmd = exec.Command("xl2tpd-control", "connect", eptInfo.сonnName)
	)

	log.G(d.ctx).Info("creating ppp options file", zap.String("ppo_opt_file", eptInfo.pppOptFile),
		zap.String("network_id", eptInfo.networkID), zap.String("endpoint_id", eptInfo.id))
	if err := ioutil.WriteFile(eptInfo.pppOptFile, []byte(pppCfg), 0644); err != nil {
		log.G(d.ctx).Error("failed to create ppp options file", zap.String("network_id", netInfo.id),
			zap.String("endpoint_id", eptInfo.id), zap.Any("config", xl2tpdCfg), zap.Error(err))
		return nil, errors.Wrapf(err, "failed to create ppp options file for network %s, config is `%s`",
			netInfo.id, xl2tpdCfg)
	}

	log.G(d.ctx).Info("adding xl2tp connection config", zap.String("network_id", eptInfo.networkID),
		zap.String("endpoint_id", eptInfo.id), zap.Any("config", xl2tpdCfg))
	if err := addCfgCmd.Run(); err != nil {
		log.G(d.ctx).Error("failed to add xl2tpd config", zap.String("network_id", netInfo.id),
			zap.String("endpoint_id", eptInfo.id), zap.Any("config", xl2tpdCfg), zap.Error(err))
		return nil, errors.Wrapf(err, "failed to add xl2tpd connection config for network %s, config is `%s`",
			netInfo.id, xl2tpdCfg)
	}

	var (
		linkEvents        = make(chan netlink.LinkUpdate)
		linkEventsStopper = make(chan struct{})
		addrEvents        = make(chan netlink.AddrUpdate)
		addrEventsStopper = make(chan struct{})
	)
	if err := netlink.LinkSubscribe(linkEvents, linkEventsStopper); err != nil {
		log.G(d.ctx).Error("failed to subscribe to netlink", zap.String("network_id", eptInfo.networkID),
			zap.String("endpoint_id", eptInfo.id), zap.Error(err))
		return nil, errors.Wrapf(err, "failed to subscribe to netlink: %s", err)
	}

	if err := netlink.AddrSubscribe(addrEvents, addrEventsStopper); err != nil {
		log.G(d.ctx).Error("failed to subscribe to netlink", zap.String("network_id", eptInfo.networkID),
			zap.String("endpoint_id", eptInfo.id), zap.Error(err))
		return nil, errors.Wrapf(err, "failed to subscribe to netlink: %s", err)
	}

	log.G(d.ctx).Info("setting up xl2tpd connection", zap.String("connection_name", eptInfo.сonnName),
		zap.String("network_id", eptInfo.networkID), zap.String("endpoint_id", eptInfo.id))
	if err := setupConnCmd.Run(); err != nil {
		log.G(d.ctx).Error("xl2tpd failed to setup connection", zap.String("network_id", netInfo.id),
			zap.Any("config", xl2tpdCfg), zap.Error(err))
		return nil, errors.Wrapf(err, "failed to add xl2tpd config for network %s, config is `%s`",
			netInfo.id, xl2tpdCfg)
	}

	var (
		linkIndex int
		inetAddr  string
	)

	linkTicker := time.NewTicker(time.Second * 5)
	for {
		var done bool

		select {
		case update := <-linkEvents:
			if update.Attrs().Name == eptInfo.pppDevName {
				linkIndex = update.Link.Attrs().Index
				done = true
			}
		case <-linkTicker.C:
			log.G(d.ctx).Error("failed to receive link update: timeout",
				zap.String("network_id", netInfo.id), zap.Any("config", xl2tpdCfg))
			return nil, errors.New("failed to receive link update: timeout")
		}

		if done {
			break
		}
	}

	linkEventsStopper <- struct{}{}

	addrTicker := time.NewTicker(time.Second * 5)
	for {
		var done bool
		select {
		case update := <-addrEvents:
			if update.LinkIndex == linkIndex {
				inetAddr = update.LinkAddress.String()
				done = true
			}
		case <-addrTicker.C:
			log.G(d.ctx).Error("failed to receive addr update: timeout",
				zap.String("network_id", netInfo.id), zap.Any("config", xl2tpdCfg))
			return nil, errors.New("failed to receive addr update: timeout")
		}

		if done {
			break
		}
	}

	addrEventsStopper <- struct{}{}

	netInfo.endpoints[eptInfo.id] = eptInfo

	return &network.CreateEndpointResponse{
		Interface: &network.EndpointInterface{Address: inetAddr},
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
	id       string
	ipv4Data []*network.IPAMData
	ipv6Data []*network.IPAMData

	lnsAddr     string
	pppUsername string
	pppPassword string

	endpoints map[string]*endpointInfo
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
	id          string
	networkID   string
	lnsAddr     string
	сonnName    string
	pppDevName  string
	pppOptFile  string
	pppUsername string
	pppPassword string
}

func (i *endpointInfo) setup() error {
	i.сonnName = i.id[:5] + "-connection"
	i.pppDevName = "ppp" + i.id[:5]
	i.pppOptFile = pppOptsDir + i.networkID[:5] + "." + i.id[:5] + ".client"

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
