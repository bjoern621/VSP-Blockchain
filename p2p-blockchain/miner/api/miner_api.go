package api

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
)

// MinerAPI provides methods to control the mining process.
type MinerAPI interface {
	StartMining(transactions []transaction.Transaction)
	StopMining()
}
