package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api"
)

type InternsalViewService struct {
	networkRegistryAPI api.NetworkInfoAPI
}

func NewInternsalViewService(networkRegistryAPI api.NetworkInfoAPI) *InternsalViewService {
	return &InternsalViewService{
		networkRegistryAPI: networkRegistryAPI,
	}
}

func (s *InternsalViewService) GetPeerRegistry() []api.PeerInfo {
	return s.networkRegistryAPI.GetPeers()
}
