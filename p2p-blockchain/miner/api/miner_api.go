package api

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
)

// MinerAPI provides methods to control the mining process.
type MinerAPI interface {
	// StartMining starts mining with the given transactions (only if mining is enabled)
	StartMining(transactions []transaction.Transaction)
	// StopMining stops the current mining process (only if mining is enabled)
	StopMining()
	// EnableMining enables mining capability
	EnableMining()
	// DisableMining disables mining capability and stops any ongoing mining
	DisableMining()
}
