package handshake

import (
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"

	"bjoernblessin.de/go-utils/util/logger"
)

// HandshakeHandler defines the interface for handling incoming connection messages.
// This interface is implemented in the core/domain layer and used by the infrastructure layer.
type HandshakeHandler interface {
	HandleVersion(peerID peer.PeerID, info VersionInfo)
	HandleVerack(peerID peer.PeerID, info VersionInfo)
	HandleAck(peerID peer.PeerID)
}

func (h *HandshakeService) HandleVersion(peerID peer.PeerID, info VersionInfo) {
	// Domain logic:
	// 1. Validate version compatibility
	// 2. Store peer info
	// 3. Send Verack back to the peer (via MessageSender interface)
	logger.Infof("Received Version from peer %s: %+v", peerID, info)
}

func (h *HandshakeService) HandleVerack(peerID peer.PeerID, info VersionInfo) {
	// Domain logic:
	// 1. Validate the verack
	// 2. Send Ack back to complete the handshake
}

func (h *HandshakeService) HandleAck(peerID peer.PeerID) {
	// Domain logic:
	// 1. Mark connection as fully established
}
