package api

import (
	"net/netip"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"
)

// RegistryEntry represents a single peer address entry from the DNS registry.
type RegistryEntry struct {
	IPAddress netip.Addr
	PeerID    peer.PeerID
}

// QueryRegistryAPI provides access to the registry for (manual) peer discovery.
type QueryRegistryAPI interface {
	// QueryRegistry queries the registry for available peer addresses.
	QueryRegistry() ([]RegistryEntry, error)
}

type queryRegistryService struct {
	querier peer.RegistryQuerier
}

// NewQueryRegistryService creates a new QueryRegistryAPI.
func NewQueryRegistryService(querier peer.RegistryQuerier) QueryRegistryAPI {
	return &queryRegistryService{
		querier: querier,
	}
}

func (s *queryRegistryService) QueryRegistry() ([]RegistryEntry, error) {
	// Note: This simplified version only returns peer IDs.
	// The full registry entry with IP addresses is available from the infrastructure layer.
	peers, err := s.querier.QueryPeers()
	if err != nil {
		return nil, err
	}

	entries := make([]RegistryEntry, 0, len(peers))
	for _, peerID := range peers {
		entries = append(entries, RegistryEntry{
			PeerID: peerID,
		})
	}

	return entries, nil
}
