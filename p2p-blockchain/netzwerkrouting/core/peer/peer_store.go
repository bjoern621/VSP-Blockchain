package peer

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"sync"
)

type PeerStore struct {
	mu    sync.RWMutex
	peers map[common.PeerId]*Peer
}

// PeerCreator is an interface for creating new peers.
type PeerCreator interface {
	NewOutboundPeer() common.PeerId
	NewInboundPeer() common.PeerId
}

func NewPeerStore() *PeerStore {
	return &PeerStore{
		peers: make(map[common.PeerId]*Peer),
	}
}

func (s *PeerStore) GetPeer(id common.PeerId) (*Peer, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	peer, exists := s.peers[id]
	return peer, exists
}

func (s *PeerStore) addPeer(peer *Peer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.peers[peer.id] = peer
}

func (s *PeerStore) RemovePeer(id common.PeerId) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.peers, id)
}

// NewInboundPeer creates a new peer for an inbound connection.
func (s *PeerStore) NewInboundPeer() common.PeerId {
	return s.NewPeer(DirectionInbound)
}

// NewOutboundPeer creates a new peer for an outbound connection.
func (s *PeerStore) NewOutboundPeer() common.PeerId {
	return s.NewPeer(DirectionOutbound)
}
