package core

import (
	"errors"
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

func (s *ConnectionEstablishmentService) ConnectTo(ip []byte, port uint16) error {
	ipAddr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return errors.New("invalid IP address format")
	}

	addrPort := netip.AddrPortFrom(ipAddr, port)
	return s.handshakeAPI.InitiateHandshake(addrPort)
}
