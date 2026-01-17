package peer

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"sync"
)

// peerStore manages the storage and retrieval of peer information.
// It primarily implements the PeerCreator and PeerRetriever interfaces that other layers define.
type peerStore struct {
	mu    sync.RWMutex
	peers map[common.PeerId]*Peer
}

func NewPeerStore() *peerStore {
	return &peerStore{
		peers: make(map[common.PeerId]*Peer),
	}
}

// GetPeer retrieves a peer by its ID.
func (s *peerStore) GetPeer(id common.PeerId) (*Peer, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	peer, exists := s.peers[id]
	return peer, exists
}

// GetAllPeers retrieves all known peers.
func (s *peerStore) GetAllPeers() []common.PeerId {
	s.mu.RLock()
	defer s.mu.RUnlock()

	peerIds := make([]common.PeerId, 0, len(s.peers))
	for k := range s.peers {
		peerIds = append(peerIds, k)
	}
	return peerIds
}

// GetAllOutboundPeers retrieves all connected outbound peers' IDs.
func (s *peerStore) GetAllOutboundPeers() []common.PeerId {
	s.mu.Lock()
	defer s.mu.Unlock()

	peerIds := make([]common.PeerId, 0)

	for k, v := range s.peers {
		if v.Direction == common.DirectionOutbound && v.State == common.StateConnected {
			peerIds = append(peerIds, k)
		}
	}

	return peerIds
}

// GetPeersWithHandshakeStarted retrieves all peers' IDs that have started the handshake process.
// These are peers that are in StateAwaitingVerack, StateAwaitingAck, or StateConnected.
func (s *peerStore) GetPeersWithHandshakeStarted() []common.PeerId {
	s.mu.RLock()
	defer s.mu.RUnlock()

	peerIds := make([]common.PeerId, 0)

	for k, v := range s.peers {
		if v.State == common.StateAwaitingVerack || v.State == common.StateAwaitingAck || v.State == common.StateConnected {
			peerIds = append(peerIds, k)
		}
	}

	return peerIds
}

// GetUnconnectedPeers retrieves peer IDs that are known but not currently connected.
// Technically, these are peers with StateNew.
func (s *peerStore) GetUnconnectedPeers() []common.PeerId {
	s.mu.RLock()
	defer s.mu.RUnlock()

	peerIds := make([]common.PeerId, 0)
	for k, v := range s.peers {
		if v.State == common.StateNew {
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

// NewPeer creates a new peer without a specified direction.
func (s *peerStore) NewPeer() common.PeerId {
	return s.newGenericPeer()
}
