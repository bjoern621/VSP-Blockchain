package api

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	corepeer "s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"
	"slices"
)

// FullInfrastructureInfo is a map from PeerID to arbitrary infrastructure data.
// The infrastructure layer is free to fill this with any data it wants.
// Callers can serialize the data to JSON or string for display.
type FullInfrastructureInfo map[common.PeerId]map[string]any

// InfrastructureInfoProvider provides access to network-level (infrastructure / grpc) information about peers.
type InfrastructureInfoProvider interface {
	// GetAllInfrastructureInfo returns all available information for all peers.
	// Returns a map from PeerID to arbitrary data that the infrastructure layer provides.
	GetAllInfrastructureInfo() FullInfrastructureInfo
}

// PeerInfo contains information about a peer from the network registry and peer store.
type PeerInfo struct {
	PeerID common.PeerId

	// PeerInfrastructureData contains arbitrary data from the infrastructure layer.
	// This can be serialized to JSON for display.
	PeerInfrastructureData map[string]any

	Version           string
	ConnectionState   common.PeerConnectionState
	SupportedServices []common.ServiceType
	LastSeen          int64
}

// NetworkInfoAPI provides access to peer information.
type NetworkInfoAPI interface {
	GetInternalPeerInfo() []PeerInfo
}

// peerRetriever is an interface for retrieving peers.
type peerRetriever interface {
	GetPeer(id common.PeerId) (corepeer.PeerData, bool)
}

type networkRegistryService struct {
	networkInfoProvider InfrastructureInfoProvider
	peerRetriever       peerRetriever
}

func NewNetworkRegistryService(networkInfoProvider InfrastructureInfoProvider, peerRetriever peerRetriever) NetworkInfoAPI {
	return &networkRegistryService{
		networkInfoProvider: networkInfoProvider,
		peerRetriever:       peerRetriever,
	}
}

func (s *networkRegistryService) GetInternalPeerInfo() []PeerInfo {
	allInfo := s.networkInfoProvider.GetAllInfrastructureInfo()

	result := make([]PeerInfo, 0, len(allInfo))

	for peerID, infraData := range allInfo {
		pInfo := PeerInfo{
			PeerID:                 peerID,
			PeerInfrastructureData: infraData,
		}

		if p, exists := s.peerRetriever.GetPeer(peerID); exists {
			p.Lock()
			pInfo.Version = p.GetVersion()
			pInfo.ConnectionState = p.GetState()

			pInfo.SupportedServices = slices.Clone(p.GetSupportedServices())
			pInfo.LastSeen = p.GetLastSeen()
			p.Unlock()
		}

		result = append(result, pInfo)
	}

	return result
}
