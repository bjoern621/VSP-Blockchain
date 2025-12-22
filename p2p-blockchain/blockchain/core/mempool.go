package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/validation"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"

	"sync"

	"bjoernblessin.de/go-utils/util/assert"
)

type Mempool struct {
	validator validation.ValidationService

	transactions map[transaction.TransactionID]transaction.Transaction
	lock         sync.Mutex
}

func NewMempool(validator validation.ValidationService) *Mempool {
	return &Mempool{
		validator:    validator,
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
	ok, err := m.validator.ValidateTransaction(&tx)
	assert.Assert(ok && err == nil, "Transaction is invalid: %v", err)

	m.lock.Lock()
	defer m.lock.Unlock()

	txId := tx.TransactionId()
	_, ok = m.transactions[txId]
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
