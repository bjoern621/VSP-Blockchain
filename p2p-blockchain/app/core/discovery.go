package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api"
)

type DiscoveryService struct {
	discoveryAPI api.DiscoveryAPI
}

func NewDiscoveryService(discoveryAPI api.DiscoveryAPI) *DiscoveryService {
	return &DiscoveryService{
		discoveryAPI: discoveryAPI,
	}
}

func (s *DiscoveryService) SendGetAddr(peerID string) error {
	pid := common.PeerId(peerID)
	s.discoveryAPI.SendGetAddr(pid)
	return nil
}

// GossipService handles periodic gossip-based peer discovery.
type GossipService struct {
	gossipAPI api.GossipAPI
}

func NewGossipService(gossipAPI api.GossipAPI) *GossipService {
	return &GossipService{
		gossipAPI: gossipAPI,
	}
}

// Start begins the gossip discovery process.
func (s *GossipService) Start() {
	s.gossipAPI.Start()
}

// Stop halts the gossip discovery process.
func (s *GossipService) Stop() {
	s.gossipAPI.Stop()
}
