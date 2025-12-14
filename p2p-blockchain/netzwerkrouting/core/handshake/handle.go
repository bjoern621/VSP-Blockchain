package handshake

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"

	"bjoernblessin.de/go-utils/util/logger"
)

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

func (h *HandshakeService) HandleVersion(peerID common.PeerId, info VersionInfo) {
	p, ok := h.peerStore.GetPeer(peerID)
	if !ok {
		logger.Warnf("unknown peer %s sent Version message", peerID)
		return
	}

	p.Lock()
	defer p.Unlock()

	if p.State != peer.StateNew {
		logger.Warnf("peer %s sent Version message in invalid state %v", peerID, p.State)
		return
	}

	if !checkVersionCompatibility(info.Version) {
		logger.Warnf("peer %s has incompatible version %s", peerID, info.Version)
		return
	}

	// Valid

	p.Version = info.Version
	p.SupportedServices = info.SupportedServices

	versionInfo := VersionInfo{
		Version:           common.VersionString,
		SupportedServices: []peer.ServiceType{peer.ServiceType_Netzwerkrouting, peer.ServiceType_BlockchainFull, peer.ServiceType_Wallet, peer.ServiceType_Miner},
	}

	p.State = peer.StateAwaitingAck

	go h.handshakeMsgSender.SendVerack(peerID, versionInfo)
}

func (h *HandshakeService) HandleVerack(peerID common.PeerId, info VersionInfo) {
	p, ok := h.peerStore.GetPeer(peerID)
	if !ok {
		logger.Warnf("unknown peer %s sent Verack message", peerID)
		return
	}

	p.Lock()
	defer p.Unlock()

	if p.State != peer.StateAwaitingVerack {
		logger.Warnf("peer %s sent Verack message in invalid state %v", peerID, p.State)
		return
	}

	if !checkVersionCompatibility(info.Version) {
		logger.Warnf("peer %s has incompatible version %s", peerID, info.Version)
		return
	}

	// Valid

	p.State = peer.StateConnected
	p.Version = info.Version
	p.SupportedServices = info.SupportedServices

	go h.handshakeMsgSender.SendAck(peerID)
}

func (h *HandshakeService) HandleAck(peerID common.PeerId) {
	p, ok := h.peerStore.GetPeer(peerID)
	if !ok {
		logger.Warnf("unknown peer %s sent Ack message", peerID)
		return
	}

	p.Lock()
	defer p.Unlock()

	if p.State != peer.StateAwaitingAck {
		logger.Warnf("peer %s sent Ack message in invalid state %v", peerID, p.State)
		return
	}

	p.State = peer.StateConnected
}
