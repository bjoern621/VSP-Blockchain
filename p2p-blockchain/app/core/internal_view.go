package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api"
)

type InternalViewService struct {
	networkInfoAPI api.NetworkInfoAPI
}

func NewInternsalViewService(networkInfoAPI api.NetworkInfoAPI) *InternalViewService {
	return &InternalViewService{
		networkInfoAPI: networkInfoAPI,
	}
}

func (svc *InternalViewService) GetInternalPeerInfo() []api.PeerInfo {
	return svc.networkInfoAPI.GetInternalPeerInfo()
}
