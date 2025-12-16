package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api"
)

type InternalViewService struct {
	networkRegistryAPI api.NetworkInfoAPI
}

func NewInternsalViewService(networkRegistryAPI api.NetworkInfoAPI) *InternalViewService {
	return &InternalViewService{
		networkRegistryAPI: networkRegistryAPI,
	}
}

func (svc *InternalViewService) GetInternalPeerInfo() []api.PeerInfo {
	return svc.networkRegistryAPI.GetInternalPeerInfo()
}
