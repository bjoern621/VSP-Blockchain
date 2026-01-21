package api

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer/discovery"
)

// DiscoveryAPI is the external API for peer discovery operations.
// Part of NetworkroutingAppAPI.
type DiscoveryAPI interface {
	// SendGetAddr sends a GetAddr request to the specified peer to request known peer addresses.
	SendGetAddr(peerID common.PeerId)
}

// discoveryAPIService implements DiscoveryAPI.
type discoveryAPIService struct {
	discoveryService *discovery.DiscoveryService
}

func NewDiscoveryAPIService(discoveryService *discovery.DiscoveryService) DiscoveryAPI {
	return &discoveryAPIService{
		discoveryService: discoveryService,
	}
}

func (s *discoveryAPIService) SendGetAddr(peerID common.PeerId) {
	s.discoveryService.SendGetAddr(peerID)
}
