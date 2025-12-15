package peer

import "bjoernblessin.de/go-utils/util/assert"

type ServiceType int

const (
	ServiceType_Netzwerkrouting ServiceType = iota
	ServiceType_BlockchainFull
	ServiceType_BlockchainSimple
	ServiceType_Wallet
	ServiceType_Miner
)

func (s ServiceType) String() string {
	switch s {
	case ServiceType_Netzwerkrouting:
		return "netzwerkrouting"
	case ServiceType_BlockchainFull:
		return "blockchain_full"
	case ServiceType_BlockchainSimple:
		return "blockchain_simple"
	case ServiceType_Wallet:
		return "wallet"
	case ServiceType_Miner:
		return "miner"
	default:
		assert.Never("unhandled ServiceType")
		return "unknown"
	}
}
