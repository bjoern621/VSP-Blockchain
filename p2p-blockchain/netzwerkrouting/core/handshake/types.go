package handshake

import (
	"net/netip"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"
)

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

// HandshakeService implements ConnectionHandler (for infrastructure) and HandshakeAPI (for api) with the actual domain logic.
type HandshakeService struct {
	handshakeInitiator HandshakeInitiator
	peerStore          *peer.PeerStore
}

// Compile-time check that HandshakeService implements specified interfaces
var _ HandshakeHandler = (*HandshakeService)(nil)
var _ api.HandshakeAPI = (*HandshakeService)(nil)

func NewHandshakeService(handshakeInitiator HandshakeInitiator, peerStore *peer.PeerStore) *HandshakeService {
	return &HandshakeService{
		handshakeInitiator: handshakeInitiator,
		peerStore:          peerStore,
	}
}
