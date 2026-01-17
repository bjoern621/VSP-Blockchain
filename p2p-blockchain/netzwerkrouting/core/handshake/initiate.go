package handshake

import (
	"fmt"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"

	"bjoernblessin.de/go-utils/util/assert"
)

// HandshakeMsgSender defines the interface for initiating a handshake with a peer.
// It is implemented by the infrastructure layer.
type HandshakeMsgSender interface {
	// SendVersion sends a Version message to the specified peer.
	SendVersion(peerID common.PeerId, info VersionInfo)
	// SendVerack sends a Verack message to the specified peer.
	SendVerack(peerID common.PeerId, info VersionInfo)
	// SendAck sends an Ack message to the specified peer.
	SendAck(peerID common.PeerId)
}

// HandshakeInitiator defines the interface for initiating handshakes with peers.
type HandshakeInitiator interface {
	// InitiateHandshake starts the handshake process with the given peer.
	InitiateHandshake(peerID common.PeerId) error
}

func (h *handshakeService) InitiateHandshake(peerID common.PeerId) error {
	p, ok := h.peerRetriever.GetPeer(peerID)
	if !ok {
		return fmt.Errorf("peer %s not found in store", peerID)
	}

	p.Lock()
	defer p.Unlock()

	if p.State != common.StateNew {
		return fmt.Errorf("cannot initiate handshake with peer %s in state %v. peer state must be StateNew", peerID, p.State)
	}

	switch p.Direction {
	case common.DirectionBoth:
		return fmt.Errorf("cannot initiate handshake with peer %s with direction Both", peerID)
	case common.DirectionInbound:
		p.Direction = common.DirectionBoth
	case common.DirectionOutbound:
		return fmt.Errorf("handshake already initiated with outbound peer %s", peerID)
	case common.DirectionUnknown:
		p.Direction = common.DirectionOutbound
	default:
		assert.Never("unhandled peer direction")
	}

	versionInfo := NewLocalVersionInfo()

	p.State = common.StateAwaitingVerack

	go h.handshakeMsgSender.SendVersion(peerID, versionInfo)

	return nil
}
