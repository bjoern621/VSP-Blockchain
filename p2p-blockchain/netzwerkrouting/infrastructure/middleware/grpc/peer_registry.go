package grpc

import (
	"net/netip"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"
)

// PeerRegistry maintains a registry of peers and their network addresses.
// It should allow representing peers by a generic ID and not relying on IP addresses internally.
type PeerRegistry struct {
	peers map[netip.AddrPort]peer.PeerID
}

func NewPeerRegistry() *PeerRegistry {
	return &PeerRegistry{
		peers: make(map[netip.AddrPort]peer.PeerID),
	}
}

func (r *PeerRegistry) GetOrCreatePeerID(addr netip.AddrPort) peer.PeerID {
	if id, exists := r.peers[addr]; exists {
		return id
	}

	peerID := peer.NewPeer()
	r.peers[addr] = peerID

	return peerID
}
