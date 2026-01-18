package disconnect

import (
	"fmt"
	"net/netip"

	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/data/peer"
)

// DisconnectService defines the interface for disconnecting/forgetting peers.
type DisconnectService interface {
	// Disconnect disconnects from a peer at the given address.
	// This involves:
	// 1. Finding the peer by address
	// 2. Verifying the peer exists in the store
	// 3. Removing from network info registry (closes gRPC connections)
	// 4. Removing from peer store
	Disconnect(addrPort netip.AddrPort) error
}

// outboundPeerResolver is implemented by infrastructure layer (NetworkInfoRegistry) to resolve
// peers by address.
type outboundPeerResolver interface {
	// GetOutboundPeer checks if a peer with the given listening address already exists.
	GetOutboundPeer(addrPort netip.AddrPort) (peerID common.PeerId, exists bool)
}

// networkInfoRemover is implemented by infrastructure layer (NetworkInfoRegistry) to remove
// peers from the network registry.
type networkInfoRemover interface {
	// RemovePeer removes a peer from the network info registry.
	// This closes any active connections.
	RemovePeer(id common.PeerId)
}

// peerRetriever is an interface for retrieving peers.
// It is implemented by peer.PeerStore.
type peerRetriever interface {
	GetPeer(id common.PeerId) (*peer.Peer, bool)
}

// peerRemover is an interface for removing peers.
// It is implemented by peer.PeerStore.
type peerRemover interface {
	RemovePeer(id common.PeerId)
}

// disconnectService implements DisconnectService with the actual domain logic.
type disconnectService struct {
	outboundPeerResolver outboundPeerResolver
	networkInfoRemover   networkInfoRemover
	peerRetriever        peerRetriever
	peerRemover          peerRemover
}

func NewDisconnectService(
	outboundPeerResolver outboundPeerResolver,
	networkInfoRemover networkInfoRemover,
	peerStore peerRetrieverAndRemover,
) DisconnectService {
	return &disconnectService{
		outboundPeerResolver: outboundPeerResolver,
		networkInfoRemover:   networkInfoRemover,
		peerRetriever:        peerStore,
		peerRemover:          peerStore,
	}
}

// peerRetrieverAndRemover combines peerRetriever and peerRemover interfaces.
// It is implemented by peer.PeerStore.
type peerRetrieverAndRemover interface {
	peerRetriever
	peerRemover
}

// Disconnect disconnects from a peer at the given address.
// This involves:
// 1. Finding the peer by address
// 2. Verifying the peer exists in the store
// 3. Removing from network info registry (closes gRPC connections)
// 4. Removing from peer store
func (s *disconnectService) Disconnect(addrPort netip.AddrPort) error {
	// Get the peer by address
	peerID, exists := s.outboundPeerResolver.GetOutboundPeer(addrPort)
	if !exists {
		return fmt.Errorf("peer not found at %s", addrPort.String())
	}

	// Verify the peer exists in the store
	_, ok := s.peerRetriever.GetPeer(peerID)
	if !ok {
		return fmt.Errorf("peer %s not found in store", peerID)
	}

	// Remove from network info registry first (closes gRPC connections)
	s.networkInfoRemover.RemovePeer(peerID)

	// Then remove from peer store
	s.peerRemover.RemovePeer(peerID)

	return nil
}
