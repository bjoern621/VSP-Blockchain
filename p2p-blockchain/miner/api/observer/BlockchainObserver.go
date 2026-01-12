package observer

import "s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"

type BlockchainObserverAPI interface {
	StartMining(transactions []transaction.Transaction)
	StopMining()
}
