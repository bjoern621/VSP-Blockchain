package disconnect

import (
	"fmt"

	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/data/peer"

	"bjoernblessin.de/go-utils/util/logger"
)

// DisconnectService defines the interface for disconnecting/forgetting peers.
type DisconnectService interface {
	// Disconnect disconnects from a peer with the given ID.
	// This involves:
	// 1. Verifying the peer exists in the store
	// 2. Removing from network info registry (closes gRPC connections)
	// 3. Removing from peer store
	Disconnect(peerID common.PeerId) error
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
	networkInfoRemover networkInfoRemover
	peerRetriever      peerRetriever
	peerRemover        peerRemover
}

func NewDisconnectService(
	networkInfoRemover networkInfoRemover,
	peerStore peerRetrieverAndRemover,
) DisconnectService {
	return &disconnectService{
		networkInfoRemover: networkInfoRemover,
		peerRetriever:      peerStore,
		peerRemover:        peerStore,
	}
}

// peerRetrieverAndRemover combines peerRetriever and peerRemover interfaces.
// It is implemented by peer.PeerStore.
type peerRetrieverAndRemover interface {
	peerRetriever
	peerRemover
}

// Disconnect disconnects from a peer with the given ID.
// This involves:
// 1. Verifying the peer exists in the store
// 2. Removing from network info registry (closes gRPC connections)
// 3. Removing from peer store
func (s *disconnectService) Disconnect(peerID common.PeerId) error {
	// Verify the peer exists in the store
	_, ok := s.peerRetriever.GetPeer(peerID)
	if !ok {
		return fmt.Errorf("peer %s not found in store", peerID)
	}

	logger.Infof("Disconnecting from peer %s", peerID)

	// Remove from network info registry first (closes gRPC connections)
	s.networkInfoRemover.RemovePeer(peerID)

	// Then remove from peer store
	s.peerRemover.RemovePeer(peerID)

	return nil
}
