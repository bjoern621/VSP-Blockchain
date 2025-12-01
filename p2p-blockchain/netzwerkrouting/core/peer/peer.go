// Package peer handles peer management of the network using generic identifiers.
package peer

import (
	"bjoernblessin.de/go-utils/util/logger"
	"github.com/google/uuid"
)

type PeerID string

type PeerConnectionState int

const (
	StateFirstSeen PeerConnectionState = iota
	StateVersionReceived
	StateVerackReceived
	StateHandshakeComplete
)

type Direction int

const (
	DirectionInbound Direction = iota
	DirectionOutbound
)

// Peer represents a peer in the network.
type Peer struct {
	ID                PeerID
	Version           string
	SupportedServices []string
	State             PeerConnectionState
	Direction         Direction
}

// NewPeer creates a new peer with a unique ID and adds it to the peer store.
// PeerConnectionState is initialized to StateFirstSeen which indicates that we just received the first message by this peer or we are trying to establish a connection.
func NewPeer(direction Direction) PeerID {
	peerID := PeerID(uuid.NewString())
	peer := &Peer{
		ID:        peerID,
		State:     0,
		Direction: direction,
	}
	peerStore.AddPeer(peer)
	logger.Debugf("new peer %v created", peerID)
	return peerID
}
