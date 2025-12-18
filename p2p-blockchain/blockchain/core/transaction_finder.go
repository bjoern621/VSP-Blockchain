package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/transaction"
)

func (b *Blockchain) IsTransactionKnown(hash block.Hash) bool {
	//TODO: Implement
	panic("not implemented")
	return true
}

func (b *Blockchain) IsTransactionKnownById(hash transaction.TransactionID) bool {
	panic("not implemented")
	return true
}
