package handshake

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"

	"bjoernblessin.de/go-utils/util/logger"
)

// errorMsgSender defines the interface for sending error/reject messages to peers.
// This allows the core layer to send reject messages without depending on the API layer.
type errorMsgSender interface {
	// SendReject sends a reject message to the specified peer
	SendReject(peerId common.PeerId, errorType int32, rejectedMessageType string, data []byte)
}

// HandshakeMsgHandler defines the interface for handling incoming connection messages.
// This interface is implemented in the core/domain layer and used by the infrastructure layer.
type HandshakeMsgHandler interface {
	HandleVersion(peerID common.PeerId, info VersionInfo)
	HandleVerack(peerID common.PeerId, info VersionInfo)
	HandleAck(peerID common.PeerId)
}

func checkVersionCompatibility(string) bool {
	return true
}

func (h *handshakeService) HandleVersion(peerID common.PeerId, info VersionInfo) {
	p, ok := h.peerRetriever.GetPeer(peerID)
	if !ok {
		logger.Warnf("[handshake_handler] unknown peer %s sent Version message", peerID)
		h.errorMsgSender.SendReject(peerID, common.ErrorTypeRejectNotConnected, "version", []byte(info.Version))
		return
	}

	p.Lock()
	defer p.Unlock()

	if p.State != common.StateNew {
		logger.Warnf("[handshake_handler] peer %s sent Version message in invalid state %v", peerID, p.State)
		h.errorMsgSender.SendReject(peerID, common.ErrorTypeRejectInvalid, "version", []byte(p.State.String()))
		return
	}

	if !checkVersionCompatibility(info.Version) {
		logger.Warnf("[handshake_handler] peer %s has incompatible version %s", peerID, info.Version)
		h.errorMsgSender.SendReject(peerID, common.ErrorTypeRejectInvalid, "version", []byte(info.Version))
		return
	}

	// Valid

	p.Version = info.Version
	p.SupportedServices = info.SupportedServices()

	versionInfo := NewLocalVersionInfo()

	p.State = common.StateAwaitingAck

	go h.handshakeMsgSender.SendVerack(peerID, versionInfo)
}

func (h *handshakeService) HandleVerack(peerID common.PeerId, info VersionInfo) {
	p, ok := h.peerRetriever.GetPeer(peerID)
	if !ok {
		logger.Warnf("[handshake_handler] unknown peer %s sent Verack message", peerID)
		h.errorMsgSender.SendReject(peerID, common.ErrorTypeRejectNotConnected, "verack", []byte(info.Version))
		return
	}

	p.Lock()
	defer p.Unlock()

	if p.State != common.StateAwaitingVerack {
		logger.Warnf("[handshake_handler] peer %s sent Verack message in invalid state %v", peerID, p.State)
		h.errorMsgSender.SendReject(peerID, common.ErrorTypeRejectInvalid, "verack", []byte(p.State.String()))
		return
	}

	if !checkVersionCompatibility(info.Version) {
		logger.Warnf("[handshake_handler] peer %s has incompatible version %s", peerID, info.Version)
		h.errorMsgSender.SendReject(peerID, common.ErrorTypeRejectInvalid, "verack", []byte(info.Version))
		return
	}

	// Valid

	p.State = common.StateConnected
	p.Version = info.Version
	p.SupportedServices = info.SupportedServices()

	go h.handshakeMsgSender.SendAck(peerID)
}

func (h *handshakeService) HandleAck(peerID common.PeerId) {
	p, ok := h.peerRetriever.GetPeer(peerID)
	if !ok {
		logger.Warnf("[handshake_handler] unknown peer %s sent Ack message", peerID)
		h.errorMsgSender.SendReject(peerID, common.ErrorTypeRejectNotConnected, "ack", []byte("unknown peer"))
		return
	}

	p.Lock()
	defer p.Unlock()

	if p.State != common.StateAwaitingAck {
		logger.Warnf("[handshake_handler] peer %s sent Ack message in invalid state %v", peerID, p.State)
		h.errorMsgSender.SendReject(peerID, common.ErrorTypeRejectInvalid, "ack", []byte(p.State.String()))
		return
	}

	p.State = common.StateConnected
}
