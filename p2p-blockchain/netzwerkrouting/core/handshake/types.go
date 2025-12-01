package handshake

import "net/netip"

type VersionInfo struct {
	Version           string
	SupportedServices []ServiceType
	ListeningEndpoint netip.AddrPort
}

type ServiceType int

const (
	ServiceType_Netzwerkrouting ServiceType = iota
	ServiceType_BlockchainFull
	ServiceType_BlockchainSimple
	ServiceType_Wallet
	ServiceType_Miner
)

// HandshakeService implements ConnectionHandler with the actual domain logic.
type HandshakeService struct {
	handshakeInitiator HandshakeInitiator
}

// Compile-time check that HandshakeService implements HandshakeHandler
var _ HandshakeHandler = (*HandshakeService)(nil)

func NewHandshakeService(handshakeInitiator HandshakeInitiator) *HandshakeService {
	return &HandshakeService{
		handshakeInitiator: handshakeInitiator,
	}
}
