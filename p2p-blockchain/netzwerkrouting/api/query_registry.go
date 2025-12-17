package api

import (
	"net/netip"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
)

// RegistryEntry represents a single peer address entry from the DNS registry.
type RegistryEntry struct {
	IPAddress netip.Addr
	PeerID    common.PeerId
}

// QueryRegistryAPI provides access to the registry for (manual) peer discovery or plain reading.
type QueryRegistryAPI interface {
	// QueryRegistry queries the registry for available peer addresses.
	QueryRegistry() ([]RegistryEntry, error)
}

type queryRegistryAPIService struct {
	querier FullRegistryQuerier
}

func NewQueryRegistryAPIService(querier FullRegistryQuerier) QueryRegistryAPI {
	return &queryRegistryAPIService{
		querier: querier,
	}
}

// Used interface for full registry queries including IP addresses.
type FullRegistryQuerier interface {
	// QueryRegistry queries the registry for available peer addresses.
	QueryFullRegistry() ([]RegistryEntry, error)
}

func (s *queryRegistryAPIService) QueryRegistry() ([]RegistryEntry, error) {
	return s.querier.QueryFullRegistry()
}
