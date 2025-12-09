package api

import (
	"net/netip"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"
)

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
	networkInfoProvider core.NetworkInfoProvider
	peerStore           *peer.PeerStore
}

func NewNetworkRegistryService(networkInfoProvider core.NetworkInfoProvider, peerStore *peer.PeerStore) NetworkInfoAPI {
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
			pInfo.ConnectionState = connectionStateToString(p.State)
			pInfo.Direction = directionToString(p.Direction)

			for _, svc := range p.SupportedServices {
				pInfo.SupportedServices = append(pInfo.SupportedServices, serviceTypeToString(svc))
			}
			p.Unlock()
		}

		result = append(result, pInfo)
	}

	return result
}

func connectionStateToString(state peer.PeerConnectionState) string {
	switch state {
	case peer.StateNew:
		return "new"
	case peer.StateAwaitingVerack:
		return "awaiting_verack"
	case peer.StateAwaitingAck:
		return "awaiting_ack"
	case peer.StateConnected:
		return "connected"
	default:
		return "unknown"
	}
}

func directionToString(dir peer.Direction) string {
	switch dir {
	case peer.DirectionInbound:
		return "inbound"
	case peer.DirectionOutbound:
		return "outbound"
	case peer.DirectionBoth:
		return "both"
	default:
		return "unknown"
	}
}

func serviceTypeToString(svc peer.ServiceType) string {
	switch svc {
	case peer.ServiceType_Netzwerkrouting:
		return "netzwerkrouting"
	case peer.ServiceType_BlockchainFull:
		return "blockchain_full"
	case peer.ServiceType_BlockchainSimple:
		return "blockchain_simple"
	case peer.ServiceType_Wallet:
		return "wallet"
	case peer.ServiceType_Miner:
		return "miner"
	default:
		return "unknown"
	}
}
