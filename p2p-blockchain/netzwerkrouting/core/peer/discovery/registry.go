package discovery

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"

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
	_, err := s.querier.QueryPeers()
	if err != nil {
		logger.Warnf("Failed to query registry for peers: %v", err)
		return
	}

	// There is nothing more to do here because the peers are created by the infrastructure layer.
}
