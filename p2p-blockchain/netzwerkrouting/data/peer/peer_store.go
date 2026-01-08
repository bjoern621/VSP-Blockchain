package peer

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"sync"
)

// peerStore manages the storage and retrieval of peer information.
// It primarily implements the PeerCreator and PeerRetriever interfaces.
type peerStore struct {
	mu    sync.RWMutex
	peers map[common.PeerId]*Peer
}

// PeerCreator is an interface for creating new peers.
type PeerCreator interface {
	// NewOutboundPeer creates a new peer for an outbound connection.
	NewOutboundPeer() common.PeerId
	// NewInboundPeer creates a new peer for an inbound connection.
	NewInboundPeer() common.PeerId
	// NewPeer creates a new peer without a specified direction.
	NewPeer() common.PeerId
}

// PeerRetriever is an interface for retrieving peers.
type PeerRetriever interface {
	// GetPeer retrieves a peer by its ID.
	GetPeer(id common.PeerId) (*Peer, bool)
	// GetAllOutboundPeers retrieves all outbound peers' IDs.
	GetAllOutboundPeers() []common.PeerId
}

func NewPeerStore() *peerStore {
	return &peerStore{
		peers: make(map[common.PeerId]*Peer),
	}
}

func (s *peerStore) GetPeer(id common.PeerId) (*Peer, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	peer, exists := s.peers[id]
	return peer, exists
}

func (s *peerStore) GetAllOutboundPeers() []common.PeerId {
	s.mu.Lock()
	defer s.mu.Unlock()

	peerIds := make([]common.PeerId, 0)

	for k, v := range s.peers {
		if v.Direction == common.DirectionOutbound {
			peerIds = append(peerIds, k)
		}
	}

	return peerIds
}

func (s *peerStore) addPeer(peer *Peer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.peers[peer.id] = peer
}

func (s *peerStore) RemovePeer(id common.PeerId) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.peers, id)
}

// NewInboundPeer creates a new peer for an inbound connection.
func (s *peerStore) NewInboundPeer() common.PeerId {
	return s.newPeer(common.DirectionInbound)
}

// NewOutboundPeer creates a new peer for an outbound connection.
func (s *peerStore) NewOutboundPeer() common.PeerId {
	return s.newPeer(common.DirectionOutbound)
}

func (s *peerStore) NewPeer() common.PeerId {
	return s.newGenericPeer()
}
