package discovery

import "s3b/vsp-blockchain/p2p-blockchain/internal/common"

// RegistryQuerier abstracts registry lookups for the core layer.
// This interface is implemented by the infrastructure layer which handles network details.
type RegistryQuerier interface {
	// QueryPeers queries the registry and returns discovered peers.
	QueryPeers() ([]common.PeerId, error)
}

// GetPeers queries the registry and creates peers for each discovered address.
// Returns the peer IDs of the discovered peers.
func (s *DiscoveryService) GetPeers(hostname string) ([]common.PeerId, error) {
	return s.querier.QueryPeers()
}
