package handshake

import (
	"fmt"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"
)

// HandshakeMsgSender defines the interface for initiating a handshake with a peer.
// It is implemented by the infrastructure layer.
type HandshakeMsgSender interface {
	// SendVersion sends a Version message to the specified peer.
	SendVersion(peerID peer.PeerID, info VersionInfo)
	// SendVerack sends a Verack message to the specified peer.
	SendVerack(peerID peer.PeerID, info VersionInfo)
	// SendAck sends an Ack message to the specified peer.
	SendAck(peerID peer.PeerID)
}

// HandshakeInitiator defines the interface for initiating handshakes with peers.
type HandshakeInitiator interface {
	// InitiateHandshake starts the handshake process with the given peer.
	InitiateHandshake(peerID peer.PeerID) error
}

func (h *handshakeService) InitiateHandshake(peerID peer.PeerID) error {
	p, ok := h.peerStore.GetPeer(peerID)
	if !ok {
		return fmt.Errorf("peer %s not found in store", peerID)
	}

	p.Lock()
	defer p.Unlock()

	if p.State != peer.StateNew {
		return fmt.Errorf("cannot initiate handshake with peer %s in state %v. peer state must be StateNew", peerID, p.State)
	}

	versionInfo := NewLocalVersionInfo()

	p.State = peer.StateAwaitingVerack

	go h.handshakeMsgSender.SendVersion(peerID, versionInfo)

	return nil
}
