package npp

import (
	"context"
	"net"

	"github.com/libp2p/go-reuseport"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/insonmnia/rendezvous"
	"github.com/sonm-io/core/util/xgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc/credentials"
)

// TODO
type Option func(o *options) error

type options struct {
	ctx       context.Context
	log       *zap.Logger
	localAddr net.Addr
	client    rendezvous.Client
}

// TODO
func newOptions(ctx context.Context) *options {
	return &options{
		ctx:       ctx,
		localAddr: &net.TCPAddr{IP: net.IPv4zero}, // TODO: IPv6 ready? Huh.
	}
}

// WithRendezvous is an option that specifies Rendezvous client settings.
//
// Without this option no intermediate server will be used for obtaining
// peer's endpoints and the entire connection establishment process will fall
// back to the old good plain TCP connection.
func WithRendezvous(addr auth.Endpoint, credentials credentials.TransportCredentials) Option {
	return func(o *options) error {
		conn, err := reuseport.Dial("tcp", "", addr.Endpoint)
		if err != nil {
			return err
		}

		client, err := rendezvous.NewRendezvousClient(o.ctx, "", credentials, xgrpc.WithConn(conn))
		if err != nil {
			return err
		}

		o.client = client
		o.localAddr = conn.LocalAddr()

		return nil
	}
}

//
func WithLogger(log *zap.Logger) Option {
	return func(o *options) error {
		o.log = log
		return nil
	}
}
