package main

import (
	"context"
	"net"
	"syscall"

	"github.com/docker/go-plugins-helpers/network"
	log "github.com/noxiouz/zapctx/ctxlog"
	"go.uber.org/zap"
)

type TincNetworkDriver struct {
	ctx context.Context
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
	return nil
}

func (t *TincNetworkDriver) AllocateNetwork(request *network.AllocateNetworkRequest) (*network.AllocateNetworkResponse, error) {
	log.G(t.ctx).Info("received AllocateNetwork request", zap.Any("request", request))
	return nil, nil
}

func (t *TincNetworkDriver) DeleteNetwork(request *network.DeleteNetworkRequest) error {
	log.G(t.ctx).Info("received DeleteNetwork request", zap.Any("request", request))
	return nil
}

func (t *TincNetworkDriver) FreeNetwork(request *network.FreeNetworkRequest) error {
	log.G(t.ctx).Info("received FreeNetwork request", zap.Any("request", request))
	return nil
}

func (t *TincNetworkDriver) CreateEndpoint(request *network.CreateEndpointRequest) (*network.CreateEndpointResponse, error) {
	log.G(t.ctx).Info("received CreateEndpoint request", zap.Any("request", request))
	return &network.CreateEndpointResponse{}, nil
	//return &network.CreateEndpointResponse{Interface: request.Interface}, nil
}

func (t *TincNetworkDriver) DeleteEndpoint(request *network.DeleteEndpointRequest) error {
	log.G(t.ctx).Info("received DeleteEndpoint request", zap.Any("request", request))
	return nil
}

func (t *TincNetworkDriver) EndpointInfo(request *network.InfoRequest) (*network.InfoResponse, error) {
	log.G(t.ctx).Info("received EndpointInfo request", zap.Any("request", request))
	val := make(map[string]string)
	val["Endpoint"] = "10.274.0.4/16"
	return &network.InfoResponse{Value: val}, nil
}

func (t *TincNetworkDriver) Join(request *network.JoinRequest) (*network.JoinResponse, error) {
	log.G(t.ctx).Info("received Join request", zap.Any("request", request))
	//return nil, nil
	return &network.JoinResponse{DisableGatewayService: true, InterfaceName: network.InterfaceName{SrcName: "netname", DstPrefix: "pidor"}}, nil
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

func main() {
	d := TincNetworkDriver{ctx: context.Background()}

	//socketPath := "/run/docker/plugins/runtime-root/plugins.moby/tinc/tinc.sock"
	socketPath := "/run/docker/plugins/tinc/tinc.sock"

	//listener, err := sockets.NewUnixSocket(socketPath, syscall.Getgid())
	syscall.Unlink(socketPath)
	l, err := net.Listen("unix", socketPath)
	if err != nil {
		log.G(context.Background()).Error("HUYNA", zap.Error(err))
		return
	}
	defer l.Close()

	if err != nil {
		log.G(context.Background()).Error("HUYNA", zap.Error(err))
		return
	}

	h := network.NewHandler(&d)
	err = h.Serve(l)
	if err != nil {
		log.G(context.Background()).Error("HUYNA", zap.Error(err))
	} else {
		log.G(context.Background()).Info("Zaebis", zap.Error(err))
	}
}
