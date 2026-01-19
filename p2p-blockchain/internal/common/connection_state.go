package common

import "bjoernblessin.de/go-utils/util/assert"

type PeerConnectionState int

const (
	StateNew PeerConnectionState = iota // StateNew is the initial state when a peer is created
	StateAwaitingVerack
	StateAwaitingAck
	StateConnected // Handshake complete
	StateHolddown  // Peer is in holddown period after disconnect, rejecting new connections
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
	case StateHolddown:
		return "holddown"
	default:
		assert.Never("unhandled PeerConnectionState")
		return "unknown"
	}
}
