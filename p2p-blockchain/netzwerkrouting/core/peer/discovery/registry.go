package discovery

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"time"

	"bjoernblessin.de/go-utils/util/assert"
	"bjoernblessin.de/go-utils/util/logger"
)

// RegistryQuerier abstracts registry lookups for the core layer.
// This interface is implemented by the infrastructure layer which handles network details.
type RegistryQuerier interface {
	// QueryPeers queries the registry and returns discovered peers.
	QueryPeers() ([]common.PeerId, error)
}

// GetPeers queries the registry and creates peers for each discovered address.
func (s *DiscoveryService) GetPeers() {
	peers, err := s.querier.QueryPeers()
	if err != nil {
		return
	}

	// There is not much to do here because the peers are created by the infrastructure layer. We just need to update their LastSeen timestamp.

	for _, peerID := range peers {
		peer, found := s.peerRetriever.GetPeer(peerID)
		assert.Assert(found, "peer should exist after registry discovery")

		if peer.State != common.StateNew {
			continue // Only update LastSeen for peers in StateNew via discovery
		}

		peer.Lock()
		peer.LastSeen = time.Now().Unix()
		peer.Unlock()
		logger.Tracef("[peer-discovery] Discovered peer from registry: %v", peerID)
	}
}
