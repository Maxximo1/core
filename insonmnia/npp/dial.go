// Client side for NPP.
// TODO: Check for credentials.

package npp

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/insonmnia/rendezvous"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/netutil"
)

type Dialer struct {
	ctx         context.Context
	client      rendezvous.Client
	localAddr   net.Addr
	maxAttempts int
	timeout     time.Duration
}

// TODO
func NewDialer(ctx context.Context, options ...Option) (*Dialer, error) {
	opts := newOptions(ctx, &net.TCPAddr{
		IP: net.IPv4(0, 0, 0, 0), // TODO: Hack.
	})

	for _, o := range options {
		if err := o(opts); err != nil {
			return nil, err
		}
	}

	return &Dialer{
		ctx:         ctx,
		client:      opts.client,
		localAddr:   opts.localAddr,
		maxAttempts: 5,
		timeout:     10 * time.Second,
	}, nil
}

func (m *Dialer) Dial(addr auth.Endpoint) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(m.ctx, m.timeout)
	defer cancel()

	addrs, err := m.rendezvous(ctx, addr)
	if err != nil {
		return nil, err
	}

	return m.punch(ctx, addrs)
}

func (m *Dialer) punch(ctx context.Context, addrs *sonm.ConnectReply) (net.Conn, error) {
	if addrs.Empty() {
		return nil, fmt.Errorf("no addresses resolved")
	}

	pending := make(chan connTuple)

	if addrs.PublicAddr.IsValid() {
		go func() {
			conn, err := m.punchAddr(ctx, addrs.PublicAddr)
			pending <- newConnTuple(conn, err)
		}()
	}

	for _, addr := range addrs.PrivateAddrs {
		go func() {
			pending <- newConnTuple(m.punchAddr(ctx, addr))
		}()
	}

	waiter := make(chan connTuple)

	go func() {
		var errs []error
		bullshit := false
		for conn := range pending {
			if conn.error != nil {
				errs = append(errs, conn.error)
				continue
			}

			if bullshit {
				conn.Close()
			} else {
				waiter <- conn
				bullshit = true
			}
		}

		waiter <- newConnTuple(nil, fmt.Errorf("failed to punch the network: all attempts has failed: %+v", errs))
	}()

	conn := <-waiter
	return conn.unwrap()
}

func (m *Dialer) punchAddr(ctx context.Context, addr *sonm.Addr) (net.Conn, error) {
	peerAddr, err := addr.IntoTCP()
	if err != nil {
		return nil, err
	}

	var conn net.Conn
	for i := 0; i < m.maxAttempts; i++ {
		conn, err = DialContext(ctx, protocol, m.localAddr.String(), peerAddr.String())
		if err == nil {
			return conn, nil
		}
	}

	return nil, fmt.Errorf("failed to connect: %s", err)
}

func (m *Dialer) rendezvous(ctx context.Context, addr auth.Endpoint) (*sonm.ConnectReply, error) {
	privateAddrs, err := m.privateAddrs()
	if err != nil {
		return nil, err
	}

	request := &sonm.ConnectRequest{
		Protocol:     protocol,
		PrivateAddrs: []*sonm.Addr{},
		ID:           addr.EthAddress.String(),
	}

	// TODO: Decompose.
	for _, addr := range privateAddrs {
		host, port, err := netutil.SplitHostPort(addr.String())
		if err != nil {
			return nil, err
		}

		request.PrivateAddrs = append(request.PrivateAddrs, &sonm.Addr{
			Protocol: protocol,
			Addr: &sonm.SocketAddr{
				Addr: host.String(),
				Port: uint32(port),
			},
		})
	}

	return m.client.Resolve(ctx, request)
}

func (m *Dialer) privateAddrs() ([]net.Addr, error) {
	return privateAddrs(m.localAddr)
}

// Close closes the dialer.
//
// Any blocked operations will be unblocked and return errors.
func (m *Dialer) Close() error {
	return m.client.Close()
}
