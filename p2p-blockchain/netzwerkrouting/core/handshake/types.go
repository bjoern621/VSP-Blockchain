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

// HandshakeService implements ConnectionHandler (for infrastructure) and HandshakeAPI (for api) with the actual domain logic.
type HandshakeService struct {
	handshakeMsgSender HandshakeMsgSender
	peerStore          *peer.PeerStore
}

func NewHandshakeService(handshakeInitiator HandshakeMsgSender, peerStore *peer.PeerStore) *HandshakeService {
	return &HandshakeService{
		handshakeMsgSender: handshakeInitiator,
		peerStore:          peerStore,
	}
}
