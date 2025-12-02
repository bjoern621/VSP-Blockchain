package handshake

import (
	"net/netip"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"

	"bjoernblessin.de/go-utils/util/logger"
)

// HandshakeInitiator defines the interface for initiating a handshake with a peer.
// It is implemented by the infrastructure layer.
type HandshakeInitiator interface {
	SendVersion(peerID peer.PeerID, info VersionInfo)
	SendVerack(peerID peer.PeerID, info VersionInfo)
	SendAck(peerID peer.PeerID)
}

// HandshakeService interface for the API layer.
type HandshakeServiceAPI interface {
	InitiateHandshake(peerID peer.PeerID)
}

func (h *HandshakeService) InitiateHandshake(peerID peer.PeerID) {
	p, ok := h.peerStore.GetPeer(peerID)
	if !ok {
		logger.Warnf("peer %s not found in store", peerID)
		return
	}

	versionInfo := VersionInfo{
		Version:           common.VersionString,
		SupportedServices: []peer.ServiceType{peer.ServiceType_Netzwerkrouting, peer.ServiceType_BlockchainFull, peer.ServiceType_Wallet, peer.ServiceType_Miner},
		ListeningEndpoint: netip.AddrPortFrom(common.P2PListeningIpAddr, common.P2PPort),
	}

	p.State = peer.StateAwaitingVerack

	h.handshakeInitiator.SendVersion(peerID, versionInfo)
}
