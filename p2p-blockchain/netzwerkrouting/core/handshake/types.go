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
	handshakeInitiator HandshakeInitiator
	peerStore          *peer.PeerStore
}

// Compile-time check that HandshakeService implements specific interfaces
var _ HandshakeHandler = (*HandshakeService)(nil)
var _ HandshakeServiceAPI = (*HandshakeService)(nil)

func NewHandshakeService(handshakeInitiator HandshakeInitiator, peerStore *peer.PeerStore) *HandshakeService {
	return &HandshakeService{
		handshakeInitiator: handshakeInitiator,
		peerStore:          peerStore,
	}
}
