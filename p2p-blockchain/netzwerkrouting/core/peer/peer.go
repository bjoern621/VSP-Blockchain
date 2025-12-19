// Package peer handles peer management of the network using generic identifiers.
package peer

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"sync"

	"bjoernblessin.de/go-utils/util/logger"
	"github.com/google/uuid"
)

// Peer represents a peer in the network.
type Peer struct {
	mu                sync.Mutex
	id                common.PeerId
	Version           string
	SupportedServices []ServiceType
	State             PeerConnectionState
	Direction         Direction
}

// newPeer creates a new peer with a unique ID and adds it to the peer store.
// PeerConnectionState is initialized to StateNew.
func (s *PeerStore) newPeer(direction Direction) common.PeerId {
	peerID := common.PeerId(uuid.NewString())
	peer := &Peer{
		id:        peerID,
		State:     StateNew,
		Direction: direction,
	}
	s.addPeer(peer)
	logger.Debugf("new peer %v created (direction: %v)", peerID, direction)
	return peerID
}

// newGenericPeer creates a new peer without a specified direction.
// Otherwise similar to newPeer.
func (s *PeerStore) newGenericPeer() common.PeerId {
	peerID := common.PeerId(uuid.NewString())
	peer := &Peer{
		id:    peerID,
		State: StateNew,
	}
	s.addPeer(peer)
	logger.Debugf("new peer %v created", peerID)
	return peerID
}

func (p *Peer) ID() common.PeerId {
	// id is immutable, no lock needed
	return p.id
}

// Lock acquires the peer's mutex. Caller must call Unlock when done.
func (p *Peer) Lock() {
	p.mu.Lock()
}

// Unlock releases the peer's mutex.
func (p *Peer) Unlock() {
	p.mu.Unlock()
}
