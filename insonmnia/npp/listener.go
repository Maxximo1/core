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
	"go.uber.org/zap"
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
	log         *zap.Logger
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
		log:         opts.log,
		client:      opts.client,
		channel:     channel,
		listener:    listener,
		maxAttempts: 30,
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
		m.log.Info("received acceptor peer", zap.Any("addr", conn.RemoteAddr()))
		return conn.Conn, conn.error
	default:
	}

	addrs, err := m.rendezvous()
	if err != nil {
		m.log.Error("fuck you remote peer endpoints", zap.Error(err))
		return nil, err
	}

	m.log.Info("received remote peer endpoints", zap.Any("addrs", addrs))

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

	m.log.Info("punch #1", zap.Any("addrs", addrs))
	pending := make(chan connTuple)

	if addrs.PublicAddr.IsValid() {
		go func() {
			conn, err := m.punchAddr(ctx, addrs.PublicAddr)
			m.log.Info("using NAT", zap.Any("addr", addrs.PublicAddr), zap.Error(err))
			pending <- newConnTuple(conn, err)
		}()
	}

	for _, addr := range addrs.PrivateAddrs {
		go func() {
			conn, err := m.punchAddr(ctx, addr)
			m.log.Info("using private address", zap.Any("addr", addr), zap.Error(err))
			pending <- newConnTuple(conn, err)
		}()
	}

	waiter := make(chan connTuple)

	go func() {
		var errs []error
		connected := false
		for conn := range pending { // TODO: Hangs here.
			if conn.error != nil {
				m.log.Info("failed to punch", zap.Error(conn.error))
				errs = append(errs, conn.error)
				continue
			}

			if connected {
				conn.Close()
			} else {
				connected = true
				waiter <- conn
			}
		}

		if !connected {
			waiter <- newConnTuple(nil, fmt.Errorf("failed to punch the network: all attempts has failed: %+v", errs))
		}
	}()

	conn := <-waiter
	m.log.Info("punched finally", zap.Stringer("addr", conn.RemoteAddr()))
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
		fmt.Printf("WTF punch: %s %s %s\n", m.Addr().String(), peerAddr.String(), err)
		if err == nil {
			return conn, nil
		}
	}

	return nil, fmt.Errorf("failed to connect: %s", err)
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
