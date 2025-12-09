package core

import (
	"net/netip"

	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api"
)

// AppService defines the core logic for the application.
type AppService interface {
	ConnectTo(ip netip.Addr, port uint16) error
	GetPeerRegistry() []api.PeerInfo
}

type appService struct {
	handshakeAPI       api.HandshakeAPI
	networkRegistryAPI api.NetworkInfoAPI
}

// NewAppService creates a new AppService.
func NewAppService(handshakeAPI api.HandshakeAPI, networkRegistryAPI api.NetworkInfoAPI) AppService {
	return &appService{
		handshakeAPI:       handshakeAPI,
		networkRegistryAPI: networkRegistryAPI,
	}
}

func (s *appService) ConnectTo(ip netip.Addr, port uint16) error {
	addrPort := netip.AddrPortFrom(ip, port)
	return s.handshakeAPI.InitiateHandshake(addrPort)
}

func (s *appService) GetPeerRegistry() []api.PeerInfo {
	return s.networkRegistryAPI.GetPeers()
}
