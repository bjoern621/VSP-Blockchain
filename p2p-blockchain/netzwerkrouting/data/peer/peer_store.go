package peer

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"sync"

	"bjoernblessin.de/go-utils/util/logger"
	"github.com/google/uuid"
)

// peerStore manages the storage and retrieval of peer information.
// It primarily implements the PeerCreator and PeerRetriever interfaces that other layers define.
type peerStore struct {
	mu    sync.RWMutex
	peers map[common.PeerId]*common.Peer
}

func NewPeerStore() *peerStore {
	return &peerStore{
		peers: make(map[common.PeerId]*common.Peer),
	}
}

// GetPeer retrieves a peer by its ID.
// Also workds for holddown peers.
func (s *peerStore) GetPeer(id common.PeerId) (*common.Peer, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	peer, exists := s.peers[id]
	return peer, exists
}

// GetAllPeers retrieves all known peers, excluding peers in holddown state.
// Holddown peers are effectively "soft-deleted" and should not be visible to most operations.
func (s *peerStore) GetAllPeers() []common.PeerId {
	s.mu.RLock()
	defer s.mu.RUnlock()

	peerIds := make([]common.PeerId, 0, len(s.peers))
	for k, v := range s.peers {
		if v.State != common.StateHolddown {
			peerIds = append(peerIds, k)
		}
	}
	return peerIds
}

// GetAllConnectedPeers retrieves all connected peers' IDs.
func (s *peerStore) GetAllConnectedPeers() []common.PeerId { // TODO
	s.mu.Lock()
	defer s.mu.Unlock()

	peerIds := make([]common.PeerId, 0)

	for k, v := range s.peers {
		if v.State == common.StateConnected {
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

// GetHolddownPeers retrieves all peers that are in holddown state.
// These are peers that have been disconnected and are awaiting cleanup.
// Used by ConnectionCheckService to check if holddown has expired.
func (s *peerStore) GetHolddownPeers() []common.PeerId {
	s.mu.RLock()
	defer s.mu.RUnlock()

	peerIds := make([]common.PeerId, 0)
	for k, v := range s.peers {
		if v.State == common.StateHolddown {
			peerIds = append(peerIds, k)
		}
	}

	return peerIds
}

func (s *peerStore) addPeer(peer *common.Peer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.peers[peer.ID()] = peer
}

func (s *peerStore) RemovePeer(id common.PeerId) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.peers, id)
}

// NewPeer creates a new peer.
func (s *peerStore) NewPeer() common.PeerId {
	return s.newGenericPeer()
}

// newPeer creates a new peer with a unique ID and adds it to the peer store.
// PeerConnectionState is initialized to StateNew.
func (s *peerStore) newGenericPeer() common.PeerId { // TODO
	peerID := common.PeerId(uuid.NewString())
	peer := common.NewPeer(peerID)
	s.addPeer(peer)
	logger.Debugf("[peer] new peer %v created", peerID)
	return peerID
}
