package api

import (
	"net/netip"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/handshake"
)

// HandshakeAPI is the external API for initiating connections.
// Part of NetworkroutingAppAPI.
type HandshakeAPI interface {
	// InitiateHandshake starts the handshake process with the given address.
	InitiateHandshake(addrPort netip.AddrPort) error
}

// OutboundPeerResolver is implemented by the infrastructure layer (NetworkInfoRegistry) to resolve
// and register peers for outbound connections.
type OutboundPeerResolver interface {
	// GetOutboundPeer checks if a peer with the given listening address already exists.
	// Used before initiating a handshake to resolve the peer ID.
	GetOutboundPeer(addrPort netip.AddrPort) (peerID common.PeerId, exists bool)
	// RegisterPeer registers a new peer with its listening endpoint in the NetworkInfoRegistry.
	// This allows the gRPC middleware to route subsequent requests to the correct peer.
	RegisterPeer(peerID common.PeerId, listeningEndpoint netip.AddrPort)
}

// peerCreator is an interface for creating new peers.
// It is implemented by the data layer's PeerStore.
type peerCreator interface {
	NewPeer() common.PeerId
}

// handshakeAPIService implements HandshakeAPI.
type handshakeAPIService struct {
	outboundPeerResolver OutboundPeerResolver
	peerCreator          peerCreator
	handshakeInitiator   handshake.HandshakeInitiator
}

func NewHandshakeAPIService(outboundPeerResolver OutboundPeerResolver, peerCreator peerCreator, handshakeService handshake.HandshakeInitiator) HandshakeAPI {
	return &handshakeAPIService{
		outboundPeerResolver: outboundPeerResolver,
		peerCreator:          peerCreator,
		handshakeInitiator:   handshakeService,
	}
}

func (s *handshakeAPIService) InitiateHandshake(addrPort netip.AddrPort) error {
	peerID, exists := s.outboundPeerResolver.GetOutboundPeer(addrPort)
	if !exists {
		peerID = s.peerCreator.NewPeer()
		s.outboundPeerResolver.RegisterPeer(peerID, addrPort)
	}

	err := s.handshakeInitiator.InitiateHandshake(peerID)
	if err != nil {
		return err
	}

	return nil
}
