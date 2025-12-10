package peer

import "bjoernblessin.de/go-utils/util/assert"

type PeerConnectionState int

const (
	StateNew PeerConnectionState = iota // StateNew is the initial state when a peer is created
	StateAwaitingVerack
	StateAwaitingAck
	StateConnected // Handshake complete
)

func (s PeerConnectionState) String() string {
	switch s {
	case StateNew:
		return "new"
	case StateAwaitingVerack:
		return "awaiting_verack"
	case StateAwaitingAck:
		return "awaiting_ack"
	case StateConnected:
		return "connected"
	default:
		assert.Never("unhandled PeerConnectionState")
		return "unknown"
	}
}
