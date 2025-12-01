package peerregistry

import (
	"net/netip"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"
)

// PeerRegistry maintains a registry of peers and their network addresses.
// It should allow representing peers by a generic ID and not relying on IP addresses internally.
type PeerRegistry struct {
	addrToPeer map[netip.AddrPort]peer.PeerID
	peerToAddr map[peer.PeerID]netip.AddrPort
	peerStore  *peer.PeerStore
}

func NewPeerRegistry(peerStore *peer.PeerStore) *PeerRegistry {
	return &PeerRegistry{
		addrToPeer: make(map[netip.AddrPort]peer.PeerID),
		peerToAddr: make(map[peer.PeerID]netip.AddrPort),
		peerStore:  peerStore,
	}
}

// GetOrCreatePeerID returns the PeerID for the given address.
// If no PeerID exists for the address, a new one is created and returned.
// Is called when receiving a connection from a peer.
func (r *PeerRegistry) GetOrCreatePeerID(addr netip.AddrPort) peer.PeerID {
	if id, exists := r.addrToPeer[addr]; exists {
		return id
	}

	peerID := r.peerStore.NewPeer(peer.DirectionInbound)
	r.addrToPeer[addr] = peerID
	r.peerToAddr[peerID] = addr

	return peerID
}

func (r *PeerRegistry) GetAddrPort(peerID peer.PeerID) (netip.AddrPort, bool) {
	addr, exists := r.peerToAddr[peerID]
	return addr, exists
}

func (r *PeerRegistry) RemovePeer(peerID peer.PeerID) {
	if addr, exists := r.peerToAddr[peerID]; exists {
		delete(r.peerToAddr, peerID)
		delete(r.addrToPeer, addr)
	}
}

func (r *PeerRegistry) AddPeer(peerID peer.PeerID, addr netip.AddrPort) {
	r.peerToAddr[peerID] = addr
	r.addrToPeer[addr] = peerID
}
