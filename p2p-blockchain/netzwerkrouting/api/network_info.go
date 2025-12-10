package api

import (
	"net/netip"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"
)

// FullNetworkInfo contains all available information about a peer.
type FullNetworkInfo struct {
	PeerID            peer.PeerID
	ListeningEndpoint netip.AddrPort
	InboundAddresses  []netip.AddrPort
	HasOutboundConn   bool
}

// NetworkInfoProvider provides access to network-level information about peers.
type NetworkInfoProvider interface {
	// GetAllNetworkInfo returns all available information for all peers.
	GetAllNetworkInfo() []FullNetworkInfo
}

// PeerInfo containes information about a peer from the network registry and peer store.
type PeerInfo struct {
	PeerID                string
	HasOutboundConnection bool
	ListeningEndpoint     netip.AddrPort
	InboundAddresses      []netip.AddrPort

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

	for _, info := range allInfo {
		pInfo := PeerInfo{
			PeerID:                string(info.PeerID),
			HasOutboundConnection: info.HasOutboundConn,
			ListeningEndpoint:     info.ListeningEndpoint,
			InboundAddresses:      info.InboundAddresses,
		}

		if p, exists := s.peerStore.GetPeer(info.PeerID); exists {
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
