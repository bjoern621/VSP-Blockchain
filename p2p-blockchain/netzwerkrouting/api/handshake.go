package api

import (
	"net/netip"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/handshake"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"

	"bjoernblessin.de/go-utils/util/logger"
)

// HandshakeAPI is the external API for initiating connections.
type HandshakeAPI interface {
	InitiateHandshake(addrPort netip.AddrPort)
}

// PeerRegistry is implemented by the infrastructure layer to manage peer lookups.
// Infrastructure provides the mapping between network addresses and peer identifiers.
type PeerRegistry interface {
	GetOrCreateOutboundPeer(addrPort netip.AddrPort) (peerID peer.PeerID, created bool)
}

// HandshakeAPIService implements HandshakeAPI.
type HandshakeAPIService struct {
	peerRegistry     PeerRegistry
	handshakeService handshake.HandshakeServiceAPI
}

var _ HandshakeAPI = (*HandshakeAPIService)(nil)

func NewHandshakeAPIService(peerRegistry PeerRegistry, handshakeService handshake.HandshakeServiceAPI) *HandshakeAPIService {
	return &HandshakeAPIService{
		peerRegistry:     peerRegistry,
		handshakeService: handshakeService,
	}
}

func (s *HandshakeAPIService) InitiateHandshake(addrPort netip.AddrPort) {
	peerID, created := s.peerRegistry.GetOrCreateOutboundPeer(addrPort)
	if !created {
		logger.Infof("already connected/connecting to %s (peer %s), skipping", addrPort, peerID)
		return
	}

	s.handshakeService.InitiateHandshake(peerID)
}
