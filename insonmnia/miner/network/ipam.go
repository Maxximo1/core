package network

import (
	"context"
	"io/ioutil"
	"os/exec"
	"time"

	"fmt"

	"github.com/docker/go-plugins-helpers/ipam"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
)

type IPAMDriver struct {
	ctx       context.Context
	counter   int
	netDriver *L2TPDriver
	endpoints map[string]*endpointInfo
}

func NewIPAMDriver(ctx context.Context, netDriver *L2TPDriver) *IPAMDriver {
	return &IPAMDriver{
		ctx:       ctx,
		netDriver: netDriver,
		endpoints: make(map[string]*endpointInfo),
	}
}

func (d *IPAMDriver) RequestPool(request *ipam.RequestPoolRequest) (*ipam.RequestPoolResponse, error) {
	log.G(d.ctx).Info("ipam: received RequestPool request", zap.Any("request", request))

	netInfo := &networkInfo{}

	for optName, optVal := range request.Options {
		switch optName {
		case "lns_addr":
			netInfo.lnsAddr = optVal
		case "ppp_username":
			netInfo.pppUsername = optVal
		case "ppp_password":
			netInfo.pppPassword = optVal
		}
	}

	netInfo.internalID = GetMD5Hash(netInfo.lnsAddr + netInfo.pppUsername + netInfo.pppPassword)

	if err := netInfo.setup(); err != nil {
		return nil, err
	}

	eptInfo := &endpointInfo{
		internalID:        fmt.Sprintf("%d_%s", d.counter, netInfo.internalID),
		networkInternalID: netInfo.internalID,
		lnsAddr:           netInfo.lnsAddr,
		pppUsername:       netInfo.pppUsername,
		pppPassword:       netInfo.pppPassword,
	}

	if err := eptInfo.setup(); err != nil {
		return nil, err
	}

	var (
		pppCfg       = eptInfo.getPppConfig()
		xl2tpdCfg    = eptInfo.getXl2tpConfig()
		addCfgCmd    = exec.Command("xl2tpd-control", "add", eptInfo.connName, xl2tpdCfg[0], xl2tpdCfg[1])
		setupConnCmd = exec.Command("xl2tpd-control", "connect", eptInfo.connName)
	)

	log.G(d.ctx).Info("creating ppp options file", zap.String("ppo_opt_file", eptInfo.pppOptFile),
		zap.String("network_id", eptInfo.networkID), zap.String("endpoint_id", eptInfo.internalID))
	if err := ioutil.WriteFile(eptInfo.pppOptFile, []byte(pppCfg), 0644); err != nil {
		log.G(d.ctx).Error("failed to create ppp options file", zap.String("network_id", netInfo.internalID),
			zap.String("endpoint_id", eptInfo.internalID), zap.Any("config", xl2tpdCfg), zap.Error(err))
		return nil, errors.Wrapf(err, "failed to create ppp options file for network %s, config is `%s`",
			netInfo.internalID, xl2tpdCfg)
	}

	log.G(d.ctx).Info("adding xl2tp connection config", zap.String("network_id", eptInfo.networkID),
		zap.String("endpoint_id", eptInfo.internalID), zap.Any("config", xl2tpdCfg))
	if err := addCfgCmd.Run(); err != nil {
		log.G(d.ctx).Error("failed to add xl2tpd config", zap.String("network_id", netInfo.internalID),
			zap.String("endpoint_id", eptInfo.internalID), zap.Any("config", xl2tpdCfg), zap.Error(err))
		return nil, errors.Wrapf(err, "failed to add xl2tpd connection config for network %s, config is `%s`",
			netInfo.internalID, xl2tpdCfg)
	}

	var (
		linkEvents        = make(chan netlink.LinkUpdate)
		linkEventsStopper = make(chan struct{})
		addrEvents        = make(chan netlink.AddrUpdate)
		addrEventsStopper = make(chan struct{})
	)
	if err := netlink.LinkSubscribe(linkEvents, linkEventsStopper); err != nil {
		log.G(d.ctx).Error("failed to subscribe to netlink", zap.String("network_id", eptInfo.networkID),
			zap.String("endpoint_id", eptInfo.internalID), zap.Error(err))
		return nil, errors.Wrapf(err, "failed to subscribe to netlink: %s", err)
	}

	if err := netlink.AddrSubscribe(addrEvents, addrEventsStopper); err != nil {
		log.G(d.ctx).Error("failed to subscribe to netlink", zap.String("network_id", eptInfo.networkID),
			zap.String("endpoint_id", eptInfo.internalID), zap.Error(err))
		return nil, errors.Wrapf(err, "failed to subscribe to netlink: %s", err)
	}

	log.G(d.ctx).Info("setting up xl2tpd connection", zap.String("connection_name", eptInfo.connName),
		zap.String("network_id", eptInfo.networkID), zap.String("endpoint_id", eptInfo.internalID))
	if err := setupConnCmd.Run(); err != nil {
		log.G(d.ctx).Error("xl2tpd failed to setup connection", zap.String("network_id", netInfo.internalID),
			zap.Any("config", xl2tpdCfg), zap.Error(err))
		return nil, errors.Wrapf(err, "failed to add xl2tpd config for network %s, config is `%s`",
			netInfo.internalID, xl2tpdCfg)
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
				zap.String("network_id", netInfo.internalID), zap.Any("config", xl2tpdCfg))
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
				zap.String("network_id", netInfo.internalID), zap.Any("config", xl2tpdCfg))
			return nil, errors.New("failed to receive addr update: timeout")
		}

		if done {
			break
		}
	}

	addrEventsStopper <- struct{}{}

	eptInfo.assignedIP = inetAddr
	netInfo.endpoint = eptInfo
	d.netDriver.networks[netInfo.internalID] = netInfo

	poolID := "fuck"

	d.endpoints[poolID] = eptInfo

	return &ipam.RequestPoolResponse{PoolID: poolID, Pool: inetAddr}, nil
}

func (d *IPAMDriver) RequestAddress(request *ipam.RequestAddressRequest) (*ipam.RequestAddressResponse, error) {
	log.G(d.ctx).Info("ipam: received RequestAddress request", zap.Any("request", request))

	if eptInfo, ok := d.endpoints[request.PoolID]; ok {
		return &ipam.RequestAddressResponse{Address: eptInfo.assignedIP}, nil
	}

	return nil, errors.Errorf("pool %s not found", request.PoolID)
}

func (d *IPAMDriver) GetCapabilities() (*ipam.CapabilitiesResponse, error) {
	log.G(d.ctx).Info("ipam: received GetCapabilities request")
	return &ipam.CapabilitiesResponse{RequiresMACAddress: false}, nil
}

func (d *IPAMDriver) GetDefaultAddressSpaces() (*ipam.AddressSpacesResponse, error) {
	log.G(d.ctx).Info("ipam: received GetDefaultAddressSpaces request")
	return &ipam.AddressSpacesResponse{}, nil
}

func (d *IPAMDriver) ReleasePool(request *ipam.ReleasePoolRequest) error {
	log.G(d.ctx).Info("ipam: received ReleasePool request", zap.Any("request", request))
	return nil
}

func (d *IPAMDriver) ReleaseAddress(request *ipam.ReleaseAddressRequest) error {
	log.G(d.ctx).Info("ipam: received ReleaseAddress request", zap.Any("request", request))
	return nil
}
