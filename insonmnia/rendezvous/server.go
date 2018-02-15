// Rendezvous protocol implementation. The other name - Bidirectional Locator.
//
// Rendezvous is a protocol aimed to achieve mutual address resolution for both
// client and server. This is especially useful where there is no guarantee
// that both peers are reachable directly, i.e. behind a NAT for example.
// The protocol allows for servers to publish their private network addresses
// while resolving their real remove address. This information is saved under
// some ID until a connection held and heartbeat frames transmitted.
// When a client wants to perform a connection it goes to this server, informs
// about its own private network addresses and starting from this point a
// rendezvous can be achieved.
// Both client and server are informed about remote peers public and private
// addresses and they may perform several attempts to create an actual p2p
// connection.
//
// By trying to connect to private addresses they can reach each other if and
// only if they're both in the same LAN.
// If a peer doesn't have private address, the connection is almost guaranteed
// to be established directly from the private peer to the public one.
// If both of them are located under different NATs, a TCP punching attempt can
// be performed.
// At last, when there is no hope, a special relay server can be used to
// forward the traffic.
//
// Servers should publish all of possibly reachable endpoints for all protocols
// they support.
// Clients should specify the desired protocol and ID for resolution.

package rendezvous

import (
	"errors"
	"math/rand"
	"net"
	"sync"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/xgrpc"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

type peerCandidate struct {
	Peer
	C chan<- Peer
}

type meeting struct {
	mu sync.Mutex
	// We allow multiple clients to be waited for servers.
	clients map[PeerID]peerCandidate
	// Also we allow the opposite: multiple servers can be registered for
	// fault tolerance.
	servers map[PeerID]peerCandidate
}

func newMeeting() *meeting {
	return &meeting{
		servers: map[PeerID]peerCandidate{},
		clients: map[PeerID]peerCandidate{},
	}
}

func (m *meeting) addServer(peer Peer, c chan<- Peer) {
	m.servers[peer.ID] = peerCandidate{Peer: peer, C: c}
}

func (m *meeting) addClient(peer Peer, c chan<- Peer) {
	m.clients[peer.ID] = peerCandidate{Peer: peer, C: c}
}

func (m *meeting) RandomServer() *peerCandidate {
	if len(m.servers) == 0 {
		return nil
	}

	var keys []PeerID
	for key := range m.servers {
		keys = append(keys, key)
	}

	v := m.servers[keys[rand.Intn(len(keys))]]
	return &v
}

func (m *meeting) RandomClient() *peerCandidate {
	if len(m.clients) == 0 {
		return nil
	}
	var keys []PeerID
	for key := range m.clients {
		keys = append(keys, key)
	}

	v := m.clients[keys[rand.Intn(len(keys))]]
	return &v
}

func (m *meeting) Servers() []Peer {
	var peers []Peer
	for _, p := range m.servers {
		peers = append(peers, p.Peer)
	}
	return peers
}

// TODO
// TODO: When resolving it's necessary to track also IP version. For example to be able not to return IPv6 when connecting socket is IPv4.
type Server struct {
	cfg    Config
	log    *zap.Logger
	server *grpc.Server

	mu sync.Mutex
	rv map[string]*meeting
}

// NewServer constructs a new rendezvous server using specified config and options.
// TODO: More.
func NewServer(cfg Config, options ...Option) (*Server, error) {
	opts := newOptions()
	for _, option := range options {
		option(opts)
	}

	server := &Server{
		cfg: cfg,
		log: opts.log,
		server: xgrpc.NewServer(
			opts.log,
			xgrpc.Credentials(opts.credentials),
			xgrpc.DefaultTraceInterceptor(),
		),
		rv: map[string]*meeting{},
	}

	server.log.Debug("configured authentication settings",
		zap.Stringer("eth", crypto.PubkeyToAddress(cfg.PrivateKey.PublicKey)),
	)

	sonm.RegisterRendezvousServer(server.server, server)
	return server, nil
}

func (s *Server) Resolve(ctx context.Context, request *sonm.ConnectRequest) (*sonm.ConnectReply, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}

	peerInfo, ok := peer.FromContext(ctx)
	if !ok {
		return nil, errors.New("no peer info provided")
	}

	s.log.Info("resolving remote peer", zap.String("id", request.ID))

	id := request.ID
	peerHandle := NewPeer(*peerInfo, request.PrivateAddrs)

	c := s.newServerWatch(id, peerHandle)
	defer s.removeServerWatch(id, peerHandle)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case p := <-c:
		return newConnectReply(p)
	}
}

func (s *Server) Publish(ctx context.Context, request *sonm.PublishRequest) (*sonm.PublishReply, error) {
	peerInfo, ok := peer.FromContext(ctx)
	if !ok {
		return nil, errors.New("no peer info provided")
	}

	ethAddr, err := auth.ExtractWalletFromContext(ctx)
	if err != nil {
		return nil, err
	}

	s.log.Info("publishing remote peer", zap.String("id", ethAddr.String()))

	id := ethAddr.String()
	peerHandle := NewPeer(*peerInfo, request.PrivateAddrs)

	c := s.newClientWatch(id, peerHandle)
	defer s.removeClientWatch(id, peerHandle)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case p := <-c:
		return newPublishReply(p)
	}
}

func (s *Server) newServerWatch(id string, peer Peer) <-chan Peer {
	c := make(chan Peer)

	s.mu.Lock()
	defer s.mu.Unlock()

	meeting, ok := s.rv[id]
	if ok {
		// Notify both sides immediately if there is match between candidates.
		if server := meeting.RandomServer(); server != nil {
			c <- server.Peer
			server.C <- peer
		} else {
			meeting.addClient(peer, c)
		}
	} else {
		meeting := newMeeting()
		meeting.addClient(peer, c)
		s.rv[id] = meeting
	}

	return c
}

func (s *Server) newClientWatch(id string, peer Peer) <-chan Peer {
	c := make(chan Peer)

	s.mu.Lock()
	defer s.mu.Unlock()

	meeting, ok := s.rv[id]
	if ok {
		if client := meeting.RandomClient(); client != nil {
			c <- client.Peer
			client.C <- peer
		} else {
			meeting.addServer(peer, c)
		}
	} else {
		meeting := newMeeting()
		meeting.addServer(peer, c)
		s.rv[id] = meeting
	}

	return c
}

func (s *Server) removeClientWatch(id string, peer Peer) {
	s.mu.Lock()
	defer s.mu.Unlock()

	candidates, ok := s.rv[id]
	if ok {
		delete(candidates.servers, peer.ID)
	}
}

func (s *Server) removeServerWatch(id string, peer Peer) {
	s.mu.Lock()
	defer s.mu.Unlock()

	candidates, ok := s.rv[id]
	if ok {
		delete(candidates.clients, peer.ID)
	}
}

func newConnectReply(peer Peer) (*sonm.ConnectReply, error) {
	addr, err := sonm.NewAddr(peer.Addr)
	if err != nil {
		return nil, err
	}

	response := &sonm.ConnectReply{
		PrivateAddrs: peer.privateAddrs,
	}

	if addr.Addr.IsPrivate() {
		response.PrivateAddrs = append(response.PrivateAddrs, addr)
	} else {
		response.PublicAddr = addr
	}

	return response, nil
}

func newPublishReply(peer Peer) (*sonm.PublishReply, error) {
	addr, err := sonm.NewAddr(peer.Addr)
	if err != nil {
		return nil, err
	}

	response := &sonm.PublishReply{
		Protocol:     peer.Addr.Network(),
		PrivateAddrs: peer.privateAddrs,
	}

	if addr.Addr.IsPrivate() {
		response.PrivateAddrs = append(response.PrivateAddrs, addr)
	} else {
		response.PublicAddr = addr
	}

	return response, nil
}

func (s *Server) Run() error {
	listener, err := net.Listen(s.cfg.Addr.Network(), s.cfg.Addr.String())
	if err != nil {
		return err
	}

	s.log.Info("rendezvous is ready to serve", zap.Stringer("endpoint", listener.Addr()))
	return s.server.Serve(listener)
}

// Stop stops the server gracefully.
//
// It stops the server from accepting new connections and RPCs and blocks
// until all the pending RPCs are finished.
func (s *Server) Stop() {
	s.log.Info("rendezvous is shutting down")
	s.server.GracefulStop()
}
