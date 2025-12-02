package api

import (
	"fmt"
	"net/netip"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/handshake"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"
)

// HandshakeAPI is the external API for initiating connections.
type HandshakeAPI interface {
	InitiateHandshake(addrPort netip.AddrPort) error
}

// PeerRegistry is implemented by the infrastructure layer to manage peer lookups and address mappings.
type PeerRegistry interface {
	GetOutboundPeer(addrPort netip.AddrPort) (peerID peer.PeerID, exists bool)
	RegisterPeer(peerID peer.PeerID, listeningEndpoint netip.AddrPort)
}

// HandshakeAPIService implements HandshakeAPI.
type HandshakeAPIService struct {
	peerRegistry     PeerRegistry
	peerCreator      peer.PeerCreator
	handshakeService handshake.HandshakeServiceAPI
}

var _ HandshakeAPI = (*HandshakeAPIService)(nil)

func NewHandshakeAPIService(peerRegistry PeerRegistry, peerCreator peer.PeerCreator, handshakeService handshake.HandshakeServiceAPI) *HandshakeAPIService {
	return &HandshakeAPIService{
		peerRegistry:     peerRegistry,
		peerCreator:      peerCreator,
		handshakeService: handshakeService,
	}
}

func (s *HandshakeAPIService) InitiateHandshake(addrPort netip.AddrPort) error {
	if peerID, exists := s.peerRegistry.GetOutboundPeer(addrPort); exists {
		return fmt.Errorf("peer %s already exists, cannot initiate handshake", peerID)
	}

	peerID := s.peerCreator.NewOutboundPeer()
	s.peerRegistry.RegisterPeer(peerID, addrPort)

	s.handshakeService.InitiateHandshake(peerID)
	return nil
}
