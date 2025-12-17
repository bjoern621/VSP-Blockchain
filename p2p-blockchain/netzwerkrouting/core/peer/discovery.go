package peer

// RegistryQuerier abstracts registry lookups for the core layer.
// This interface is implemented by the infrastructure layer which handles network details.
type RegistryQuerier interface {
	// QueryPeers queries the registry and returns discovered peers.
	QueryPeers() ([]PeerID, error)
}

// DiscoveryService provides peer discovery functionality.
type DiscoveryService struct {
	querier     RegistryQuerier
	peerCreator PeerCreator
}

// NewDiscoveryService creates a new DiscoveryService.
func NewDiscoveryService(querier RegistryQuerier, peerCreator PeerCreator) *DiscoveryService {
	return &DiscoveryService{
		querier:     querier,
		peerCreator: peerCreator,
	}
}

// GetPeers queries the registry and creates peers for each discovered address.
// Returns the peer IDs of the discovered peers.
func (s *DiscoveryService) GetPeers(hostname string) ([]PeerID, error) {
	return s.querier.QueryPeers()
}
