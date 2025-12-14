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

// ParseServiceType converts a string like "wallet" or "blockchain_full" into a ServiceType.
// Returns ok=false for unknown values.
func ParseServiceType(raw string) (service ServiceType, ok bool) {
	switch raw {
	case "netzwerkrouting":
		return ServiceType_Netzwerkrouting, true
	case "blockchain_full":
		return ServiceType_BlockchainFull, true
	case "blockchain_simple":
		return ServiceType_BlockchainSimple, true
	case "wallet":
		return ServiceType_Wallet, true
	case "miner":
		return ServiceType_Miner, true
	default:
		return 0, false
	}
}
