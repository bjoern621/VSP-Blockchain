package handshake

import (
	"net/netip"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"

	"bjoernblessin.de/go-utils/util/logger"
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
	InitiateHandshake(peerID peer.PeerID)
}

func (h *HandshakeService) InitiateHandshake(peerID peer.PeerID) {
	p, ok := h.peerStore.GetPeer(peerID)
	if !ok {
		logger.Warnf("peer %s not found in store", peerID)
		return
	}

	p.Lock()
	defer p.Unlock()

	if p.State != peer.StateNew {
		logger.Warnf("cannot initiate handshake with peer %s in state %v. peer state must be StateNew", peerID, p.State)
		return
	}

	versionInfo := VersionInfo{
		Version:           common.VersionString,
		SupportedServices: []peer.ServiceType{peer.ServiceType_Netzwerkrouting, peer.ServiceType_BlockchainFull, peer.ServiceType_Wallet, peer.ServiceType_Miner},
		ListeningEndpoint: netip.AddrPortFrom(common.P2PListeningIpAddr(), common.P2PPort()),
	}

	p.State = peer.StateAwaitingVerack

	go h.handshakeMsgSender.SendVersion(peerID, versionInfo)
}
