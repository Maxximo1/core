// This package is responsible for server-side of NAT Punching Protocol.
// TODO: Check for reuseport available. If not - do not try to punch the NAT.

package npp

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/libp2p/go-reuseport"
	"github.com/sonm-io/core/insonmnia/rendezvous"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/netutil"
)

type connTuple struct {
	net.Conn
	error
}

func newConnTuple(conn net.Conn, err error) connTuple {
	return connTuple{conn, err}
}

func (m *connTuple) unwrap() (net.Conn, error) {
	return m.Conn, m.error
}

// Options are: rendezvous server, private IPs usage, protocol (?), relay server(s) if any.
// TODO: What to do if the connection between us and the RV is broken?
type Listener struct {
	ctx         context.Context
	client      rendezvous.Client
	channel     chan connTuple
	listener    net.Listener
	maxAttempts int
	timeout     time.Duration
}

func NewListener(ctx context.Context, addr string, options ...Option) (net.Listener, error) {
	listener, err := reuseport.Listen(protocol, addr)
	if err != nil {
		return nil, err
	}

	opts := newOptions(ctx, listener.Addr())

	for _, o := range options {
		if err := o(opts); err != nil {
			return nil, err
		}
	}

	channel := make(chan connTuple, 1)

	m := &Listener{
		ctx:         ctx,
		client:      opts.client,
		channel:     channel,
		listener:    listener,
		maxAttempts: 3,
		timeout:     10 * time.Second,
	}

	go m.listen()

	return m, nil
}

//
// This method will firstly check whether there are pending sockets in the
// listener, returning immediately if so.
// Then an attempt to communicate with the Rendezvous server occurs by
// publishing server's ID to check if there are someone wanted to connect with
// us.
// Simultaneously additional sockets are constructed ... TODO
func (m *Listener) Accept() (net.Conn, error) {
	// Check for acceptor channel, if there is a connection - return immediately.
	select {
	case conn := <-m.channel:
		return conn.Conn, conn.error
	default:
	}

	addrs, err := m.rendezvous()
	if err != nil {
		return nil, err
	}

	// Here the race begins! We're simultaneously trying to connect to ALL
	// provided endpoints with a reasonable timeout. The first winner will
	// be the champion, while others die in agony. Life is cruel.
	conn, err := m.punch(addrs)
	if err != nil {
		return nil, err
	}

	// TODO: At last when there is no hope, use relay server.
	return conn, nil
}

func (m *Listener) rendezvous() (*sonm.PublishReply, error) {
	if m.client == nil {
		return sonm.EmptyPublishReply(), nil
	}

	request := &sonm.PublishRequest{
		PrivateAddrs: nil,
	}

	privateAddrs, err := m.privateAddrs()
	if err != nil {
		return nil, err
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

	return m.client.Publish(m.ctx, request)
}

func (m *Listener) punch(addrs *sonm.PublishReply) (net.Conn, error) {
	if addrs.Empty() {
		return nil, fmt.Errorf("no addresses resolved")
	}

	ctx, cancel := context.WithTimeout(context.Background(), m.timeout)
	defer cancel()

	pending := make(chan connTuple)
	defer close(pending)

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

func (m *Listener) punchAddr(ctx context.Context, addr *sonm.Addr) (net.Conn, error) {
	peerAddr, err := addr.IntoTCP()
	if err != nil {
		return nil, err
	}

	var conn net.Conn
	for i := 0; i < m.maxAttempts; i++ {
		conn, err = DialContext(ctx, protocol, m.Addr().String(), peerAddr.String())
		if err == nil {
			break
		}
	}

	return conn, fmt.Errorf("failed to connect: %s", err)
}

func (m *Listener) Close() error {
	var errs []error

	if err := m.listener.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := m.client.Close(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to close listener: %+v", errs)
	} else {
		return nil
	}
}

func (m *Listener) listen() error {
	for {
		conn, err := m.listener.Accept()
		m.channel <- connTuple{conn, err}
		if err != nil {
			return err
		}
	}
}

// PrivateAddrs collects and returns private addresses of a network interfaces
// the listening socket bind on.
func (m *Listener) privateAddrs() ([]net.Addr, error) {
	return privateAddrs(m.Addr())
}

func (m *Listener) Addr() net.Addr {
	return m.listener.Addr()
}
