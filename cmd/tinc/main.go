package main

import (
	"context"
	"net"
	"syscall"

	"github.com/docker/go-plugins-helpers/network"
	log "github.com/noxiouz/zapctx/ctxlog"
	sonmnet "github.com/sonm-io/core/insonmnia/miner/network"
	"go.uber.org/zap"
)

func main() {
	d := sonmnet.NewTincNetworkDriver(context.Background())

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

	h := network.NewHandler(d)
	err = h.Serve(l)
	if err != nil {
		log.G(context.Background()).Error("HUYNA", zap.Error(err))
	} else {
		log.G(context.Background()).Info("Zaebis", zap.Error(err))
	}
}
