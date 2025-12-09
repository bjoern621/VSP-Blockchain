package core

import (
	"net/netip"

	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api"
)

type ConnectionEstablishmentService struct {
	handshakeAPI api.HandshakeAPI
}

func NewConnectionEstablishmentService(handshakeAPI api.HandshakeAPI) *ConnectionEstablishmentService {
	return &ConnectionEstablishmentService{
		handshakeAPI: handshakeAPI,
	}
}

func (s *ConnectionEstablishmentService) ConnectTo(ip netip.Addr, port uint16) error {
	addrPort := netip.AddrPortFrom(ip, port)
	return s.handshakeAPI.InitiateHandshake(addrPort)
}
