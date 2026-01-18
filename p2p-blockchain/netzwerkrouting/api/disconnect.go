package api

import (
	"fmt"
	"net/netip"

	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/disconnect"
)

// DisconnectAPI is external API for disconnecting from peers.
// Disconnecting means forgetting a peer by removing it from peer store and network info registry.
type DisconnectAPI interface {
	// Disconnect disconnects from a peer at the given address.
	// This involves:
	// - Closing any gRPC connections
	// - Removing peer from network info registry
	// - Removing peer from peer store
	Disconnect(addrPort netip.AddrPort) error
}

// disconnectAPIService implements DisconnectAPI.
type disconnectAPIService struct {
	outboundPeerResolver OutboundPeerResolver
	disconnectService    disconnect.DisconnectService
}

func NewDisconnectAPIService(
	outboundPeerResolver OutboundPeerResolver,
	disconnectService disconnect.DisconnectService,
) DisconnectAPI {
	return &disconnectAPIService{
		outboundPeerResolver: outboundPeerResolver,
		disconnectService:    disconnectService,
	}
}

func (s *disconnectAPIService) Disconnect(addrPort netip.AddrPort) error {
	// Resolve peer by address
	peerID, exists := s.outboundPeerResolver.GetOutboundPeer(addrPort)
	if !exists {
		return fmt.Errorf("peer not found at %s", addrPort.String())
	}

	// Delegate to core service
	return s.disconnectService.Disconnect(peerID)
}
