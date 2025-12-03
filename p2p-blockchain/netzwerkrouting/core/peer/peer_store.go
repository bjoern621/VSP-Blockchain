package peer

type PeerStore struct {
	peers map[PeerID]*Peer
}

// PeerCreator is an interface for creating new peers.
type PeerCreator interface {
	NewOutboundPeer() PeerID
	NewInboundPeer() PeerID
}

var _ PeerCreator = (*PeerStore)(nil)

func NewPeerStore() *PeerStore {
	return &PeerStore{
		peers: make(map[PeerID]*Peer),
	}
}

func (s *PeerStore) GetPeer(id PeerID) (*Peer, bool) {
	peer, exists := s.peers[id]
	return peer, exists
}

func (s *PeerStore) addPeer(peer *Peer) {
	s.peers[peer.id] = peer
}

func (s *PeerStore) RemovePeer(id PeerID) {
	delete(s.peers, id)
}

// NewInboundPeer creates a new peer for an inbound connection.
func (s *PeerStore) NewInboundPeer() PeerID {
	return s.NewPeer(DirectionInbound)
}

// NewOutboundPeer creates a new peer for an outbound connection.
func (s *PeerStore) NewOutboundPeer() PeerID {
	return s.NewPeer(DirectionOutbound)
}
