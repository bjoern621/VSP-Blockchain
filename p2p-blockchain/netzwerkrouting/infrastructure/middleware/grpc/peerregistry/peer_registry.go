package peerregistry

import (
	"net/netip"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"

	"bjoernblessin.de/go-utils/util/assert"
	"bjoernblessin.de/go-utils/util/logger"
	"google.golang.org/grpc"
)

// PeerRegistry maintains a registry of peers and their network addresses.
// It should allow representing peers by a generic ID and not relying on IP addresses internally.
type PeerRegistry struct {
	addrToPeer      map[netip.AddrPort]peer.PeerID
	peerToAddr      map[peer.PeerID]netip.AddrPort
	peerStore       *peer.PeerStore
	peerConnections map[peer.PeerID]*grpc.ClientConn // gRPC connections to peers
}

func NewPeerRegistry(peerStore *peer.PeerStore) *PeerRegistry {
	return &PeerRegistry{
		addrToPeer:      make(map[netip.AddrPort]peer.PeerID),
		peerToAddr:      make(map[peer.PeerID]netip.AddrPort),
		peerStore:       peerStore,
		peerConnections: make(map[peer.PeerID]*grpc.ClientConn),
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
		err := r.peerConnections[peerID].Close()
		if err != nil {
			logger.Warnf("couldn't close connection to peer %s: %v", peerID, err)
		}
		delete(r.peerConnections, peerID)
	}
}

// AddPeer adds a mapping between the given PeerID and address.
// It should be called when establishing an outbound connection to a peer.
func (r *PeerRegistry) AddPeer(peerID peer.PeerID, addr netip.AddrPort, conn *grpc.ClientConn) {
	r.peerToAddr[peerID] = addr
	r.addrToPeer[addr] = peerID
	r.peerConnections[peerID] = conn
}

func (r *PeerRegistry) GetConnection(peerID peer.PeerID) (*grpc.ClientConn, bool) {
	conn, exists := r.peerConnections[peerID]
	return conn, exists
}

// AddConnection adds a gRPC connection for the given PeerID.
// Asserts that no connection for the PeerID already exists.
func (r *PeerRegistry) AddConnection(peerID peer.PeerID, conn *grpc.ClientConn) {
	_, exists := r.peerToAddr[peerID]
	assert.Assert(exists, "no address found for peer %s", peerID)
	_, exists = r.peerConnections[peerID]
	assert.Assert(!exists, "connection for peer %s already exists", peerID)

	r.peerConnections[peerID] = conn
}

// GetAllEntries returns all peer ID to address mappings for debugging.
func (r *PeerRegistry) GetAllEntries() map[peer.PeerID]netip.AddrPort {
	result := make(map[peer.PeerID]netip.AddrPort, len(r.peerToAddr))
	for k, v := range r.peerToAddr {
		result[k] = v
	}
	return result
}
