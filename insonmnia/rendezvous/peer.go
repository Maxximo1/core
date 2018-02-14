package rendezvous

import (
	"github.com/sonm-io/core/proto"
	"google.golang.org/grpc/peer"
)

type Peer struct {
	peer.Peer
	privateAddrs []*sonm.Addr
}

func NewPeer(peerInfo peer.Peer, privateAddrs []*sonm.Addr) Peer {
	return Peer{peerInfo, privateAddrs}
}
