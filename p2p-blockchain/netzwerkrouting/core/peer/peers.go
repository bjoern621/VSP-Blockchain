package peer

type PeerStore struct {
	peers map[PeerID]*Peer
}

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
