package handshake

import (
	"net/netip"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"
)

// HandshakeInitiator defines the interface for initiating a handshake with a peer.
// It is implemented by the infrastructure layer.
type HandshakeInitiator interface {
	SendVersion(peerID peer.PeerID, info VersionInfo, addrPort netip.AddrPort)
	SendVerack(peerID peer.PeerID, info VersionInfo)
	SendAck(peerID peer.PeerID)
}

func (h *HandshakeService) InitiateHandshake(addrPort netip.AddrPort) {
	peerID := h.peerStore.NewPeer(peer.DirectionOutbound)

	versionInfo := VersionInfo{
		Version:           common.VersionString,
		SupportedServices: []peer.ServiceType{peer.ServiceType_Netzwerkrouting, peer.ServiceType_BlockchainFull, peer.ServiceType_Wallet, peer.ServiceType_Miner},
		ListeningEndpoint: netip.AddrPortFrom(common.P2PListeningIpAddr, common.P2PPort),
	}

	h.handshakeInitiator.SendVersion(peerID, versionInfo, addrPort)
}
