package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"

	"sync"
)

type Mempool struct {
	transactions map[transaction.TransactionID]transaction.Transaction
	lock         sync.Mutex
}

func NewMempool() *Mempool {
	return &Mempool{
		transactions: make(map[transaction.TransactionID]transaction.Transaction),
	}
}

func (m *Mempool) IsKnownTransactionHash(hash common.Hash) bool {
	txId := getTransactionIdFromHash(hash)
	_, ok := m.transactions[txId]
	return ok
}

func (m *Mempool) IsKnownTransactionId(txId transaction.TransactionID) bool {
	_, ok := m.transactions[txId]
	return ok
}

func (m *Mempool) AddTransaction(tx transaction.Transaction) (isNew bool) {
	m.lock.Lock()
	defer m.lock.Unlock()

	txId := tx.TransactionId()
	_, ok := m.transactions[txId]
	if !ok {
		m.transactions[txId] = tx
	}

	return !ok
}

func getTransactionIdFromHash(hash common.Hash) transaction.TransactionID {
	var transactionId transaction.TransactionID
	copy(transactionId[:], hash[:])
	return transactionId
}
