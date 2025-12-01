package peer

var peerStore *PeerStore

type PeerStore struct {
	peers map[PeerID]*Peer
}

func init() {
	peerStore = newPeerStore()
}

func newPeerStore() *PeerStore {
	return &PeerStore{
		peers: make(map[PeerID]*Peer),
	}
}

func (s *PeerStore) GetPeer(id PeerID) (*Peer, bool) {
	peer, exists := s.peers[id]
	return peer, exists
}

func (s *PeerStore) AddPeer(peer *Peer) {
	s.peers[peer.ID] = peer
}

func (s *PeerStore) RemovePeer(id PeerID) {
	delete(s.peers, id)
}
