package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
)

func (b *Blockchain) IsTransactionKnown(hash common.Hash) bool {
	//TODO: Implement
	return true
}

func (b *Blockchain) IsTransactionKnownById(hash transaction.TransactionID) bool {
	return true
}
