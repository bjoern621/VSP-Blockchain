// Package peer handles peer management of the network using generic identifiers.
package common

import (
	"sync"

	mapset "github.com/deckarep/golang-set/v2"
)

// Peer represents a peer in the network.
// Peer fields can be modified if the Lock() and Unlock() methods are used.
type Peer struct {
	mu                sync.Mutex
	id                PeerId
	Version           string
	SupportedServices []ServiceType
	State             PeerConnectionState
	// LastSeen is a Unix timestamp indicating the last time the peer was seen active.
	// Seen active means, that a heartbeat message was received from the peer.
	// It's not updated on every interaction with the peer,
	// instead it's updated on discovery (gossip or registry) and heartbeat messages.
	// Discovery updates only if the peer is in StateNew.
	// Heartbeat messages update LastSeen only in StateConnected.
	// This way, we can detect unreachable peers and put them into holddown.
	LastSeen int64
	// HolddownStartTime is a Unix timestamp indicating when the peer entered holddown state.
	// A peer in holddown rejects new connections for a cooldown period (typically 15 minutes).
	// After the holddown period expires, the peer is permanently removed from the store.
	// Zero value indicates the peer is not in holddown.
	HolddownStartTime int64
	// AddrsSentTo tracks PeerIds whose addresses have been sent to this peer.
	// Prevents sending the same address twice to the same recipient.
	AddrsSentTo mapset.Set[PeerId]
}

func NewPeer(id PeerId) *Peer {
	return &Peer{
		id:                id,
		State:             StateNew,
		AddrsSentTo:       mapset.NewSet[PeerId](),
		SupportedServices: make([]ServiceType, 0),
	}
}

func (p *Peer) ID() PeerId {
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
