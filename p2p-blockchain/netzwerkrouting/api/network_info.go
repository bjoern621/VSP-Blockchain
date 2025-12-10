package api

import (
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"
)

// FullNetworkInfo is a map from PeerID to arbitrary infrastructure data.
// The infrastructure layer is free to fill this with any data it wants.
// Callers can serialize the data to JSON or string for display.
type FullNetworkInfo map[peer.PeerID]any

// NetworkInfoProvider provides access to network-level (infrastructure / grpc) information about peers.
type NetworkInfoProvider interface {
	// GetAllNetworkInfo returns all available information for all peers.
	// Returns a map from PeerID to arbitrary data that the infrastructure layer provides.
	GetAllNetworkInfo() FullNetworkInfo
}

// PeerInfo contains information about a peer from the network registry and peer store.
type PeerInfo struct {
	// InfrastructureData contains arbitrary data from the infrastructure layer.
	// This can be serialized to JSON for display.
	InfrastructureData any

	Version           string
	ConnectionState   string
	Direction         string
	SupportedServices []string
}

// NetworkInfoAPI provides access to peer information.
type NetworkInfoAPI interface {
	GetPeers() []PeerInfo
}

type networkRegistryService struct {
	networkInfoProvider NetworkInfoProvider
	peerStore           *peer.PeerStore
}

func NewNetworkRegistryService(networkInfoProvider NetworkInfoProvider, peerStore *peer.PeerStore) NetworkInfoAPI {
	return &networkRegistryService{
		networkInfoProvider: networkInfoProvider,
		peerStore:           peerStore,
	}
}

func (s *networkRegistryService) GetPeers() []PeerInfo {
	allInfo := s.networkInfoProvider.GetAllNetworkInfo()

	result := make([]PeerInfo, 0, len(allInfo))

	for peerID, infraData := range allInfo {
		pInfo := PeerInfo{
			InfrastructureData: infraData,
		}

		if p, exists := s.peerStore.GetPeer(peerID); exists {
			p.Lock()
			pInfo.Version = p.Version
			pInfo.ConnectionState = p.State.String()
			pInfo.Direction = p.Direction.String()

			for _, svc := range p.SupportedServices {
				pInfo.SupportedServices = append(pInfo.SupportedServices, svc.String())
			}
			p.Unlock()
		}

		result = append(result, pInfo)
	}

	return result
}
