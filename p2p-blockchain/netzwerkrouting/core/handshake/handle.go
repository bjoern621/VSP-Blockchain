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

	if p.State != peer.StateFirstSeen {
		logger.Warnf("peer %s sent Version message in invalid state %v", peerID, p.State)
		return
	}

	p.State = peer.StateVersionReceived
	p.Version = info.Version
	p.SupportedServices = info.SupportedServices

	logger.Infof("peer: %v", p)

	versionInfo := VersionInfo{
		Version:           common.VersionString,
		SupportedServices: []peer.ServiceType{peer.ServiceType_Netzwerkrouting, peer.ServiceType_BlockchainFull, peer.ServiceType_Wallet, peer.ServiceType_Miner},
		ListeningEndpoint: netip.AddrPortFrom(common.P2PListeningIpAddr, common.P2PPort),
	}

	h.handshakeInitiator.SendVerack(peerID, versionInfo, info.ListeningEndpoint)
}

func (h *HandshakeService) HandleVerack(peerID peer.PeerID, info VersionInfo) {
	// Domain logic:
	// 1. Validate the verack
	// 2. Send Ack back to complete the handshake
	logger.Infof("Received Verack from peer %s: %+v", peerID, info)
}

func (h *HandshakeService) HandleAck(peerID peer.PeerID) {
	// Domain logic:
	// 1. Mark connection as fully established
}
