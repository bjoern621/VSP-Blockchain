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
