package api

import (
	"fmt"
	"net/netip"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/handshake"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"
)

// HandshakeAPI is the external API for initiating connections.
type HandshakeAPI interface {
	// InitiateHandshake starts the handshake process with the given address.
	InitiateHandshake(addrPort netip.AddrPort) error
}

// OutboundPeerResolver is implemented by the infrastructure layer (NetworkInfoRegistry) to resolve
// and register peers for outbound connections.
type OutboundPeerResolver interface {
	// GetOutboundPeer checks if a peer with the given listening address already exists.
	// Used before initiating a handshake to avoid duplicate connections.
	GetOutboundPeer(addrPort netip.AddrPort) (peerID peer.PeerID, exists bool)
	// RegisterPeer registers a new peer with its listening endpoint in the NetworkInfoRegistry.
	// This allows the gRPC middleware to route subsequent requests to the correct peer.
	RegisterPeer(peerID peer.PeerID, listeningEndpoint netip.AddrPort)
}

// HandshakeAPIService implements HandshakeAPI.
type HandshakeAPIService struct {
	outboundPeerResolver OutboundPeerResolver
	peerCreator          peer.PeerCreator
	handshakeService     handshake.HandshakeServiceAPI
}

var _ HandshakeAPI = (*HandshakeAPIService)(nil)

func NewHandshakeAPIService(outboundPeerResolver OutboundPeerResolver, peerCreator peer.PeerCreator, handshakeService handshake.HandshakeServiceAPI) *HandshakeAPIService {
	return &HandshakeAPIService{
		outboundPeerResolver: outboundPeerResolver,
		peerCreator:          peerCreator,
		handshakeService:     handshakeService,
	}
}

func (s *HandshakeAPIService) InitiateHandshake(addrPort netip.AddrPort) error {
	if peerID, exists := s.outboundPeerResolver.GetOutboundPeer(addrPort); exists {
		return fmt.Errorf("peer %s already exists for address %s", peerID, addrPort)
	}

	peerID := s.peerCreator.NewOutboundPeer()
	s.outboundPeerResolver.RegisterPeer(peerID, addrPort)

	s.handshakeService.InitiateHandshake(peerID)
	return nil
}
