package api

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
)

// ErrorMsgSenderAPI defines methods to send error/reject messages to peers.
// Used by blockchain, handshake, and other services to signal errors in received messages back to the sender.
// Implemented by infrastructure layer to handle serialization and network transmission of reject messages.
type ErrorMsgSenderAPI interface {
	// SendReject sends a reject message to the specified peer
	SendReject(peerId common.PeerId, errorType int32, rejectedMessageType string, data []byte)
}
