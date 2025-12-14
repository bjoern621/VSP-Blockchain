// Package peer handles peer management of the network using generic identifiers.
package peer

import (
	"sync"

	"bjoernblessin.de/go-utils/util/logger"
	"github.com/google/uuid"
)

type PeerID string

// Peer represents a peer in the network.
type Peer struct {
	mu                sync.Mutex
	id                PeerID
	Version           string
	SupportedServices []ServiceType
	State             PeerConnectionState
	Direction         Direction
}

// NewPeer creates a new peer with a unique ID and adds it to the peer store.
// PeerConnectionState is initialized to StateNew.
func (s *PeerStore) NewPeer(direction Direction) PeerID {
	peerID := PeerID(uuid.NewString())
	peer := &Peer{
		id:        peerID,
		State:     StateNew,
		Direction: direction,
	}
	s.addPeer(peer)
	logger.Debugf("new peer %v created (direction: %v)", peerID, direction)
	return peerID
}

func (p *Peer) ID() PeerID {
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
