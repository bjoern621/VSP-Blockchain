package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/transaction"
)

type Mempool struct {
	transactions map[transaction.TransactionID]transaction.Transaction
}

func NewMempool() *Mempool {
	return &Mempool{
		transactions: make(map[transaction.TransactionID]transaction.Transaction),
	}
}

func (m *Mempool) IsKnownTransaction(hash block.Hash) bool {
	txId := getTransactionIdFromHash(hash)
	_, ok := m.transactions[txId]
	return ok
}

func (m *Mempool) AddTransaction(tx transaction.Transaction) {
	txId := tx.Hash()
	_, ok := m.transactions[txId]
	if !ok {
		m.transactions[txId] = tx
	}
}

func getTransactionIdFromHash(hash block.Hash) transaction.TransactionID {
	var transactionId transaction.TransactionID
	copy(transactionId[:], hash[:])
	return transactionId
}
