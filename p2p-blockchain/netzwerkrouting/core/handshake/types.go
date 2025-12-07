package handshake

import (
	"net/netip"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"
)

type VersionInfo struct {
	Version           string
	SupportedServices []peer.ServiceType
	ListeningEndpoint netip.AddrPort
}

// HandshakeService implements HandshakeMsgHandler (for infrastructure) and HandshakeInitiator (for api) with the actual domain logic.
type HandshakeService struct {
	handshakeMsgSender HandshakeMsgSender
	peerStore          *peer.PeerStore
}

func NewHandshakeService(handshakeMsgSender HandshakeMsgSender, peerStore *peer.PeerStore) *HandshakeService {
	return &HandshakeService{
		handshakeMsgSender: handshakeMsgSender,
		peerStore:          peerStore,
	}
}
