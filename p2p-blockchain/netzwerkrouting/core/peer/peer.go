// Package peer handles peer management of the network using generic identifiers.
package peer

import (
	"bjoernblessin.de/go-utils/util/logger"
	"github.com/google/uuid"
)

type PeerID string

type PeerConnectionState int

const (
	StateNew PeerConnectionState = iota // StateNew is the initial state when a peer is created
	StateAwaitingVerack
	StateAwaitingAck
	StateConnected // Handshake complete
)

type Direction int

const (
	DirectionInbound Direction = iota
	DirectionOutbound
	DirectionBoth
)

type ServiceType int

const (
	ServiceType_Netzwerkrouting ServiceType = iota
	ServiceType_BlockchainFull
	ServiceType_BlockchainSimple
	ServiceType_Wallet
	ServiceType_Miner
)

// Peer represents a peer in the network.
type Peer struct {
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
	return p.id
}
