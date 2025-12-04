// Package networkinfo provides the infrastructure layer for network-level peer management.
// It maintains a mapping between network-level information (IP addresses + ports) and peer identifiers from the domain layer.
package networkinfo

import (
	"net/netip"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"
	"slices"

	"bjoernblessin.de/go-utils/util/assert"
	"bjoernblessin.de/go-utils/util/logger"
	"google.golang.org/grpc"
)

// NetworkInfoEntry holds network-level information for a peer.
type NetworkInfoEntry struct {
	ListeningEndpoint netip.AddrPort   // The port we can reach them on (from VersionInfo)
	InboundAddresses  []netip.AddrPort // Inbound ports we've seen from this peer (from gRPC context)
	OutboundConn      *grpc.ClientConn // Our gRPC connection to the peer
}

// NetworkInfoRegistry maintains a registry of peers and their network addresses.
// It allows representing peers by a generic ID and links:
// - Listening endpoint (reachable address from VersionInfo)
// - Inbound addresses (ephemeral ports from gRPC context)
// - Outbound gRPC connection
// One PeerID represents one real remote node.
type NetworkInfoRegistry struct {
	listeningEndpointToPeer map[netip.AddrPort]peer.PeerID
	inboundAddrToPeer       map[netip.AddrPort]peer.PeerID
	networkInfoEntries      map[peer.PeerID]*NetworkInfoEntry
}

func NewNetworkInfoRegistry() *NetworkInfoRegistry {
	return &NetworkInfoRegistry{
		listeningEndpointToPeer: make(map[netip.AddrPort]peer.PeerID),
		inboundAddrToPeer:       make(map[netip.AddrPort]peer.PeerID),
		networkInfoEntries:      make(map[peer.PeerID]*NetworkInfoEntry),
	}
}

// GetPeerIDByAddr looks up a peer by address and port.
// Searches both listening endpoints and inbound addresses.
func (r *NetworkInfoRegistry) GetPeerIDByAddr(addr netip.AddrPort) (peer.PeerID, bool) {
	if id, exists := r.listeningEndpointToPeer[addr]; exists {
		return id, true
	}

	if id, exists := r.inboundAddrToPeer[addr]; exists {
		return id, true
	}

	return "", false
}

// GetPeerIDByAddrs looks up a peer by listening endpoint or inbound address.
// Returns the PeerID and true if found, empty string and false otherwise.
func (r *NetworkInfoRegistry) GetPeerIDByAddrs(inboundAddr netip.AddrPort, listeningEndpoint netip.AddrPort) (peer.PeerID, bool) {
	hasInbound := inboundAddr != netip.AddrPort{}
	hasListening := listeningEndpoint != netip.AddrPort{}
	assert.Assert(hasInbound || hasListening, "at least one of inboundAddr or listeningEndpoint must be provided")

	// Try to find existing peer by listening endpoint
	if hasListening {
		if peerID, exists := r.listeningEndpointToPeer[listeningEndpoint]; exists {
			return peerID, true
		}
	}

	// Try to find existing peer by inbound address
	if hasInbound {
		if peerID, exists := r.inboundAddrToPeer[inboundAddr]; exists {
			return peerID, true
		}
	}

	return "", false
}

// RegisterPeer registers an existing peerID and a listening endpoint.
func (r *NetworkInfoRegistry) RegisterPeer(peerID peer.PeerID, listeningEndpoint netip.AddrPort) {
	_, exists := r.networkInfoEntries[peerID]
	assert.Assert(!exists, "peer %s already registered", peerID)

	entry := &NetworkInfoEntry{
		ListeningEndpoint: listeningEndpoint,
	}
	r.networkInfoEntries[peerID] = entry
	if listeningEndpoint != (netip.AddrPort{}) {
		r.listeningEndpointToPeer[listeningEndpoint] = peerID
	}

	logger.Debugf("registered peer %s in network info registry infrastructure: listening=%s", peerID, listeningEndpoint)
}

// AddInboundAddress adds an inbound address to a peer's list if not already present.
// The peer must already exist.
func (r *NetworkInfoRegistry) AddInboundAddress(peerID peer.PeerID, addr netip.AddrPort) {
	entry, exists := r.networkInfoEntries[peerID]
	assert.Assert(exists, "network info entry must exist for peer %s", peerID)

	if slices.Contains(entry.InboundAddresses, addr) {
		return
	}

	entry.InboundAddresses = append(entry.InboundAddresses, addr)
	r.inboundAddrToPeer[addr] = peerID
}

// SetListeningEndpoint sets the listening endpoint for an existing peer.
func (r *NetworkInfoRegistry) SetListeningEndpoint(peerID peer.PeerID, listeningEndpoint netip.AddrPort) {
	entry, exists := r.networkInfoEntries[peerID]
	assert.Assert(exists, "network info entry must exist for peer %s", peerID)

	if entry.ListeningEndpoint == listeningEndpoint {
		return
	}

	// Remove old mapping if exists
	if entry.ListeningEndpoint != (netip.AddrPort{}) {
		delete(r.listeningEndpointToPeer, entry.ListeningEndpoint)
	}

	entry.ListeningEndpoint = listeningEndpoint
	if listeningEndpoint != (netip.AddrPort{}) {
		r.listeningEndpointToPeer[listeningEndpoint] = peerID
	}
}

// SetConnection sets the outbound gRPC connection for an existing peer.
func (r *NetworkInfoRegistry) SetConnection(peerID peer.PeerID, conn *grpc.ClientConn) {
	entry, exists := r.networkInfoEntries[peerID]
	assert.Assert(exists, "network info entry must exist for peer %s", peerID)

	assert.Assert(entry.OutboundConn == nil, "outbound connection already set for peer %s", peerID)

	entry.OutboundConn = conn
}

// GetConnection returns the outbound gRPC connection for a peer.
func (r *NetworkInfoRegistry) GetConnection(peerID peer.PeerID) (*grpc.ClientConn, bool) {
	entry, exists := r.networkInfoEntries[peerID]
	if !exists {
		return nil, false
	}
	return entry.OutboundConn, entry.OutboundConn != nil
}

// GetListeningEndpoint returns the listening endpoint for a peer.
func (r *NetworkInfoRegistry) GetListeningEndpoint(peerID peer.PeerID) (netip.AddrPort, bool) {
	entry, exists := r.networkInfoEntries[peerID]
	if !exists {
		return netip.AddrPort{}, false
	}
	return entry.ListeningEndpoint, entry.ListeningEndpoint != (netip.AddrPort{})
}

// GetOutboundPeer looks up a peer by address for outbound connections.
// Returns the PeerID and true if found, empty string and false otherwise.
func (r *NetworkInfoRegistry) GetOutboundPeer(addrPort netip.AddrPort) (peer.PeerID, bool) {
	return r.GetPeerIDByAddr(addrPort)
}

// DEBUG:

// FullNetworkInfo contains all available information about a peer.
type FullNetworkInfo struct {
	PeerID            peer.PeerID
	ListeningEndpoint netip.AddrPort
	InboundAddresses  []netip.AddrPort
	HasOutboundConn   bool
}

// GetAllNetworkInfo returns all available information for all peers.
func (r *NetworkInfoRegistry) GetAllNetworkInfo() []FullNetworkInfo {
	result := make([]FullNetworkInfo, 0, len(r.networkInfoEntries))
	for peerID, entry := range r.networkInfoEntries {
		info := FullNetworkInfo{
			PeerID:            peerID,
			ListeningEndpoint: entry.ListeningEndpoint,
			InboundAddresses:  entry.InboundAddresses,
			HasOutboundConn:   entry.OutboundConn != nil,
		}
		result = append(result, info)
	}
	return result
}
