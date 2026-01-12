package discovery

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/data/peer"
)

// RegistryQuerier abstracts registry lookups for the core layer.
// This interface is implemented by the infrastructure layer which handles network details.
type RegistryQuerier interface {
	// QueryPeers queries the registry and returns discovered peers.
	QueryPeers() ([]common.PeerId, error)
}

// DiscoveryService provides peer discovery functionality.
// This includes (1) querying a registry for peers, (2) asking neighbors for their known peers as well as (3) keeping track of active peers through heartbeats.
type DiscoveryService struct {
	querier       RegistryQuerier
	peerCreator   peer.PeerCreator
	addrMsgSender AddrMsgSender
	peerRetriever PeerRetriever
}

// NewDiscoveryService creates a new DiscoveryService.
func NewDiscoveryService(querier RegistryQuerier, peerCreator peer.PeerCreator, addrMsgSender AddrMsgSender, peerRetriever PeerRetriever) *DiscoveryService {
	return &DiscoveryService{
		querier:       querier,
		peerCreator:   peerCreator,
		addrMsgSender: addrMsgSender,
		peerRetriever: peerRetriever,
	}
}

// GetPeers queries the registry and creates peers for each discovered address.
// Returns the peer IDs of the discovered peers.
func (s *DiscoveryService) GetPeers(hostname string) ([]common.PeerId, error) {
	return s.querier.QueryPeers()
}

// PeerRetriever is an interface for retrieving peers.
type PeerRetriever interface {
	// GetAllPeers retrieves all peers.
	GetAllPeers() []common.PeerId
	// GetPeer retrieves a peer by its ID.
	GetPeer(id common.PeerId) (*peer.Peer, bool)
}

// AddrMsgSender defines the interface for sending addr messages.
type AddrMsgSender interface {
	// SendAddr sends an addr message containing a list of peer addresses to the specified peer.
	SendAddr(peerID common.PeerId, addrs []PeerAddress)
} // TODO
