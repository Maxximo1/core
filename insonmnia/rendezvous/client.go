package rendezvous

import (
	"context"

	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/xgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Client extends the generated gRPC client allowing to obtain server's
// address and to close the underlying connection.
type Client interface {
	sonm.RendezvousClient
	Close() error
}

type client struct {
	sonm.RendezvousClient
	conn *grpc.ClientConn
}

// NewRendezvousClient constructs a new rendezvous client.
func NewRendezvousClient(ctx context.Context, addr string, credentials credentials.TransportCredentials, opts ...grpc.DialOption) (Client, error) {
	conn, err := xgrpc.NewWalletAuthenticatedClient(ctx, credentials, addr, opts...)
	if err != nil {
		return nil, err
	}

	return &client{sonm.NewRendezvousClient(conn), conn}, nil
}

// Close closes the connection freeing the associated resources.
func (m *client) Close() error {
	return m.conn.Close()
}
