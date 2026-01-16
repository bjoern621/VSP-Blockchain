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
