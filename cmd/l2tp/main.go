package main

import (
	"context"
	"net"
	"syscall"

	"github.com/docker/go-plugins-helpers/ipam"
	"github.com/docker/go-plugins-helpers/network"
	log "github.com/noxiouz/zapctx/ctxlog"
	sonmnet "github.com/sonm-io/core/insonmnia/miner/network"
	"go.uber.org/zap"
)

func main() {
	go startIPAM()
	go startL2tp()

	select {}
}

func startL2tp() {
	var (
		d          = sonmnet.NewL2TPDriver(context.Background())
		socketPath = "/run/docker/plugins/l2tp.sock"
	)

	syscall.Unlink(socketPath)

	l, err := net.Listen("unix", socketPath)
	if err != nil {
		log.G(context.Background()).Error("Failed to listen", zap.Error(err))
		return
	}
	defer l.Close()

	h := network.NewHandler(d)
	if err := h.Serve(l); err != nil {
		log.G(context.Background()).Error("Failed to serve", zap.Error(err))
	}
}

func startIPAM() {
	var (
		d          = sonmnet.NewIPAMDriver(context.Background())
		socketPath = "/run/docker/plugins/ipam.sock"
	)

	syscall.Unlink(socketPath)

	l, err := net.Listen("unix", socketPath)
	if err != nil {
		log.G(context.Background()).Error("Failed to listen", zap.Error(err))
		return
	}
	defer l.Close()

	h := ipam.NewHandler(d)
	if err := h.Serve(l); err != nil {
		log.G(context.Background()).Error("Failed to serve", zap.Error(err))
	}
}
