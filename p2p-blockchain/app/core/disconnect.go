package core

import (
	"errors"
	"net/netip"

	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api"
)

type DisconnectService struct {
	disconnectAPI api.DisconnectAPI
}

func NewDisconnectService(disconnectAPI api.DisconnectAPI) *DisconnectService {
	return &DisconnectService{
		disconnectAPI: disconnectAPI,
	}
}

func (s *DisconnectService) Disconnect(ip []byte, port uint16) error {
	ipAddr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return errors.New("invalid IP address format")
	}

	addrPort := netip.AddrPortFrom(ipAddr, port)
	return s.disconnectAPI.Disconnect(addrPort)
}
