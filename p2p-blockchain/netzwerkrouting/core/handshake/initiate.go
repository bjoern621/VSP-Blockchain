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
	peerID := peer.NewPeer()

	versionInfo := VersionInfo{
		Version:           common.VersionString,
		SupportedServices: []ServiceType{ServiceType_Netzwerkrouting, ServiceType_BlockchainFull, ServiceType_Wallet, ServiceType_Miner},
		ListeningEndpoint: netip.AddrPortFrom(common.P2PListeningIpAddr, common.P2PPort),
	}

	h.handshakeInitiator.SendVersion(peerID, versionInfo, addrPort)
}
