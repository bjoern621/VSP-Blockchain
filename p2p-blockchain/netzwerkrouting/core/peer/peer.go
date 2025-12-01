// Package peer handles peer management of the network using generic identifiers.
package peer

import (
	"net/netip"

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

type VersionInfo struct {
	Version           string
	SupportedServices []string
	ListeningEndpoint netip.AddrPort
}

// Peer represents a peer in the network.
type Peer struct {
	ID          PeerID
	VersionInfo VersionInfo
	State       PeerConnectionState
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
