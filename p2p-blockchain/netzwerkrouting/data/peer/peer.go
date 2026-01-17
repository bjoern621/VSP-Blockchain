// Package peer handles peer management of the network using generic identifiers.
package peer

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"sync"

	"bjoernblessin.de/go-utils/util/logger"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
)

// Peer represents a peer in the network.
// Peer fields can be modified if the Lock() and Unlock() methods are used.
type Peer struct {
	mu                sync.Mutex
	id                common.PeerId
	Version           string
	SupportedServices []common.ServiceType
	State             common.PeerConnectionState
	Direction         common.Direction
	// LastSeen is a Unix timestamp indicating the last time the peer was seen active.
	// Seen active means, that a heartbeat message was received from the peer.
	// It's not updated on every interaction with the peer,
	// instead it's updated on discovery (gossip or registry) and heartbeat messages.
	LastSeen int64
	// AddrsSentTo tracks PeerIds whose addresses have been sent to this peer.
	// Prevents sending the same address twice to the same recipient.
	AddrsSentTo mapset.Set[common.PeerId]
}

// newPeer creates a new peer with a unique ID and adds it to the peer store.
// PeerConnectionState is initialized to StateNew.
func (s *peerStore) newPeer(direction common.Direction) common.PeerId {
	peerID := common.PeerId(uuid.NewString())
	peer := &Peer{
		id:          peerID,
		State:       common.StateNew,
		Direction:   direction,
		AddrsSentTo: mapset.NewSet[common.PeerId](),
	}
	s.addPeer(peer)
	logger.Debugf("new peer %v created (direction: %v)", peerID, direction)
	return peerID
}

// newGenericPeer creates a new peer without a specified direction.
// Otherwise similar to newPeer.
func (s *peerStore) newGenericPeer() common.PeerId {
	peerID := common.PeerId(uuid.NewString())
	peer := &Peer{
		id:          peerID,
		State:       common.StateNew,
		AddrsSentTo: mapset.NewSet[common.PeerId](),
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
