package handshake

import (
	"net/netip"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
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
	logger.Infof("Received Version from peer %s: %+v", peerID, info)

	p, ok := h.peerStore.GetPeer(peerID)
	if !ok {
		logger.Warnf("unknown peer %s sent Version message", peerID)
		return
	}

	if p.State != peer.StateNew {
		logger.Warnf("peer %s sent Version message in invalid state %v", peerID, p.State)
		return
	}

	p.Version = info.Version
	p.SupportedServices = info.SupportedServices

	logger.Infof("peer: %v", p)

	versionInfo := VersionInfo{
		Version:           common.VersionString,
		SupportedServices: []peer.ServiceType{peer.ServiceType_Netzwerkrouting, peer.ServiceType_BlockchainFull, peer.ServiceType_Wallet, peer.ServiceType_Miner},
		ListeningEndpoint: netip.AddrPortFrom(common.P2PListeningIpAddr, common.P2PPort),
	}

	p.State = peer.StateAwaitingAck

	h.handshakeInitiator.SendVerack(peerID, versionInfo)
}

func (h *HandshakeService) HandleVerack(peerID peer.PeerID, info VersionInfo) {
	logger.Infof("Received Verack from peer %s: %+v", peerID, info)

	p, ok := h.peerStore.GetPeer(peerID)
	if !ok {
		logger.Warnf("unknown peer %s sent Verack message", peerID)
		return
	}

	if p.State != peer.StateAwaitingVerack {
		logger.Warnf("peer %s sent Verack message in invalid state %v", peerID, p.State)
		return
	}

	p.State = peer.StateConnected
	p.Version = info.Version
	p.SupportedServices = info.SupportedServices

	h.handshakeInitiator.SendAck(peerID)
}

func (h *HandshakeService) HandleAck(peerID peer.PeerID) {
	logger.Infof("Received Ack from peer %s", peerID)

	p, ok := h.peerStore.GetPeer(peerID)
	if !ok {
		logger.Warnf("unknown peer %s sent Ack message", peerID)
		return
	}

	if p.State != peer.StateAwaitingAck {
		logger.Warnf("peer %s sent Ack message in invalid state %v", peerID, p.State)
		return
	}

	p.State = peer.StateConnected
}
