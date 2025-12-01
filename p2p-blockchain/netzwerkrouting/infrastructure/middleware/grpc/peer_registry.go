package grpc

import (
	"net/netip"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"
)

// PeerRegistry maintains a registry of peers and their network addresses.
// It should allow representing peers by a generic ID and not relying on IP addresses internally.
type PeerRegistry struct {
	addrToPeer map[netip.AddrPort]peer.PeerID
	peerToAddr map[peer.PeerID]netip.AddrPort
}

func NewPeerRegistry() *PeerRegistry {
	return &PeerRegistry{
		addrToPeer: make(map[netip.AddrPort]peer.PeerID),
		peerToAddr: make(map[peer.PeerID]netip.AddrPort),
	}
}

func (r *PeerRegistry) GetOrCreatePeerID(addr netip.AddrPort) peer.PeerID {
	if id, exists := r.addrToPeer[addr]; exists {
		return id
	}

	peerID := peer.NewPeer()
	r.addrToPeer[addr] = peerID
	r.peerToAddr[peerID] = addr

	return peerID
}

func (r *PeerRegistry) GetAddrPort(peerID peer.PeerID) (netip.AddrPort, bool) {
	addr, exists := r.peerToAddr[peerID]
	return addr, exists
}
