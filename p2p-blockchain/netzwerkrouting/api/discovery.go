package api

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer/discovery"
)

// DiscoveryAPI is the external API for peer discovery operations.
type DiscoveryAPI interface {
	// SendGetAddr sends a GetAddr request to the specified peer to request known peer addresses.
	SendGetAddr(peerID common.PeerId)
}

// GossipAPI is the external API for gossip-based peer discovery.
type GossipAPI interface {
	// Start begins the periodic gossip discovery.
	Start()
	// Stop halts the periodic gossip discovery.
	Stop()
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

// gossipAPIService implements GossipAPI.
type gossipAPIService struct {
	gossipDiscoveryService *discovery.GossipDiscoveryService
}

func NewGossipAPIService(gossipDiscoveryService *discovery.GossipDiscoveryService) GossipAPI {
	return &gossipAPIService{
		gossipDiscoveryService: gossipDiscoveryService,
	}
}

func (s *gossipAPIService) Start() {
	s.gossipDiscoveryService.Start()
}

func (s *gossipAPIService) Stop() {
	s.gossipDiscoveryService.Stop()
}
