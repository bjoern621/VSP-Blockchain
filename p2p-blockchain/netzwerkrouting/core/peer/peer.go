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

// Peer represents a peer in the network.
type Peer struct {
	ID                PeerID
	Version           string
	SupportedServices []string
	State             PeerConnectionState
}

func NewPeer() PeerID {
	peerID := PeerID(uuid.NewString())
	peer := &Peer{
		ID:    peerID,
		State: 0,
	}
	peerStore.AddPeer(peer)
	logger.Debugf("new peer %v created", peerID)
	return peerID
}
