// Package networkinfo provides the infrastructure layer for network-level peer management.
// It maintains a mapping between network-level information (IP addresses + ports) and peer identifiers from the domain layer.
package networkinfo

import (
	"net/netip"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/data/peer"
	"slices"
	"sync"

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
	mu                      sync.RWMutex
	listeningEndpointToPeer map[netip.AddrPort]common.PeerId
	inboundAddrToPeer       map[netip.AddrPort]common.PeerId
	networkInfoEntries      map[common.PeerId]*NetworkInfoEntry
	peerCreator             peer.PeerCreator
}

func NewNetworkInfoRegistry(peerCreator peer.PeerCreator) *NetworkInfoRegistry {
	return &NetworkInfoRegistry{
		listeningEndpointToPeer: make(map[netip.AddrPort]common.PeerId),
		inboundAddrToPeer:       make(map[netip.AddrPort]common.PeerId),
		networkInfoEntries:      make(map[common.PeerId]*NetworkInfoEntry),
		peerCreator:             peerCreator,
	}
}

// getPeerIDByAddr looks up a peer by address and port.
// Searches both listening endpoints and inbound addresses.
func (r *NetworkInfoRegistry) getPeerIDByAddr(addr netip.AddrPort) (common.PeerId, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if id, exists := r.listeningEndpointToPeer[addr]; exists {
		return id, true
	}

	if id, exists := r.inboundAddrToPeer[addr]; exists {
		return id, true
	}

	return "", false
}

// RegisterPeer registers an existing peerID and a listening endpoint.
func (r *NetworkInfoRegistry) RegisterPeer(peerID common.PeerId, listeningEndpoint netip.AddrPort) {
	r.mu.Lock()
	defer r.mu.Unlock()

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

// GetOrRegisterPeer atomically looks up a peer by addresses, or registers a new one if not found.
// Returns the peerID and true if the peer already existed, or the new peerID and false if created.
// The registered peer will have NO direction assigned yet.
func (r *NetworkInfoRegistry) GetOrRegisterPeer(inboundAddr netip.AddrPort, listeningEndpoint netip.AddrPort) common.PeerId {
	r.mu.Lock()
	defer r.mu.Unlock()

	hasInbound := inboundAddr != netip.AddrPort{}
	hasListening := listeningEndpoint != netip.AddrPort{}
	assert.Assert(hasInbound || hasListening, "at least one of inboundAddr or listeningEndpoint must be provided")

	// Try to find existing peer by listening endpoint
	if hasListening {
		if peerID, exists := r.listeningEndpointToPeer[listeningEndpoint]; exists {
			return peerID
		}
	}

	// Try to find existing peer by inbound address
	if hasInbound {
		if peerID, exists := r.inboundAddrToPeer[inboundAddr]; exists {
			return peerID
		}
	}

	// Not found, create new peer
	peerID := r.peerCreator.NewPeer()
	entry := &NetworkInfoEntry{
		ListeningEndpoint: listeningEndpoint,
	}
	r.networkInfoEntries[peerID] = entry
	if hasListening {
		r.listeningEndpointToPeer[listeningEndpoint] = peerID
	}

	logger.Debugf("registered new peer %s in network info registry: listening=%s", peerID, listeningEndpoint)
	return peerID
}

// AddInboundAddress adds an inbound address to a peer's list if not already present.
// The peer must already exist.
func (r *NetworkInfoRegistry) AddInboundAddress(peerID common.PeerId, addr netip.AddrPort) {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.networkInfoEntries[peerID]
	assert.Assert(exists, "network info entry must exist for peer %s", peerID)

	if slices.Contains(entry.InboundAddresses, addr) {
		return
	}

	entry.InboundAddresses = append(entry.InboundAddresses, addr)
	r.inboundAddrToPeer[addr] = peerID
}

// SetListeningEndpoint sets the listening endpoint for an existing peer.
func (r *NetworkInfoRegistry) SetListeningEndpoint(peerID common.PeerId, listeningEndpoint netip.AddrPort) {
	r.mu.Lock()
	defer r.mu.Unlock()

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
func (r *NetworkInfoRegistry) SetConnection(peerID common.PeerId, conn *grpc.ClientConn) {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.networkInfoEntries[peerID]
	assert.Assert(exists, "network info entry must exist for peer %s", peerID)

	assert.Assert(entry.OutboundConn == nil, "outbound connection already set for peer %s", peerID)

	entry.OutboundConn = conn
}

// GetConnection returns the outbound gRPC connection for a peer.
func (r *NetworkInfoRegistry) GetConnection(peerID common.PeerId) (*grpc.ClientConn, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, exists := r.networkInfoEntries[peerID]
	if !exists {
		return nil, false
	}
	return entry.OutboundConn, entry.OutboundConn != nil
}

// GetListeningEndpoint returns the listening endpoint for a peer.
func (r *NetworkInfoRegistry) GetListeningEndpoint(peerID common.PeerId) (netip.AddrPort, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, exists := r.networkInfoEntries[peerID]
	if !exists {
		return netip.AddrPort{}, false
	}
	return entry.ListeningEndpoint, entry.ListeningEndpoint != (netip.AddrPort{})
}

// GetOutboundPeer looks up a peer by address for outbound connections.
// Returns the PeerID and true if found, empty string and false otherwise.
func (r *NetworkInfoRegistry) GetOutboundPeer(addrPort netip.AddrPort) (common.PeerId, bool) {
	return r.getPeerIDByAddr(addrPort)
}

// GetAllInfrastructureInfo implements the InfrastructureInfoProvider interface from api.
func (r *NetworkInfoRegistry) GetAllInfrastructureInfo() api.FullInfrastructureInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(api.FullInfrastructureInfo, len(r.networkInfoEntries))
	for peerID, entry := range r.networkInfoEntries {
		result[peerID] = map[string]any{
			"peerID":            string(peerID),
			"listeningEndpoint": entry.ListeningEndpoint.String(),
			"inboundAddresses":  formatAddrPortsAsAny(entry.InboundAddresses),
			"hasOutboundConn":   entry.OutboundConn != nil,
		}
	}
	return result
}

// formatAddrPortsAsAny converts AddrPorts to []any for structpb compatibility.
func formatAddrPortsAsAny(addrs []netip.AddrPort) []any {
	result := make([]any, len(addrs))
	for i, addr := range addrs {
		result[i] = addr.String()
	}
	return result
}
