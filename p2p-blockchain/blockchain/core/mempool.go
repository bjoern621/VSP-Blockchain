package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/validation"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"

	"sync"

	"bjoernblessin.de/go-utils/util/assert"
)

type Mempool struct {
	validator validation.ValidationAPI

	transactions map[transaction.TransactionID]transaction.Transaction
	lock         sync.Mutex
}

func NewMempool(validator validation.ValidationAPI) *Mempool {
	return &Mempool{
		validator:    validator,
		transactions: make(map[transaction.TransactionID]transaction.Transaction),
	}
}

// IsKnownTransactionHash returns true if the transaction with the given Hash is known to the mempool.
func (m *Mempool) IsKnownTransactionHash(hash common.Hash) bool {
	txId := getTransactionIdFromHash(hash)
	_, ok := m.transactions[txId]
	return ok
}

// IsKnownTransactionId returns true if the transaction with the given ID is known to the mempool.
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

// Remove removes all transactions from the mempool that are included in the given block hashes.
// Also removes transactions that conflict with confirmed transactions (spend the same UTXOs).
// Finally, re-validates all remaining transactions in the mempool.
func (m *Mempool) Remove(blockHashes []common.Hash) {
	m.lock.Lock()
	defer m.lock.Unlock()

	// Track all transaction IDs to remove
	toRemove := make(map[transaction.TransactionID]bool)

	// Collect all transactions from blocks (placeholder - would need block access)
	// For now, this marks transactions for removal
	for txId := range m.transactions {
		// Re-validate each transaction
		tx := m.transactions[txId]
		ok, _ := m.validator.ValidateTransaction(&tx)
		if !ok {
			// Transaction is no longer valid (UTXOs spent, conflicts with confirmed tx)
			toRemove[txId] = true
		}
	}

	// Remove invalid transactions
	for txId := range toRemove {
		delete(m.transactions, txId)
	}
}

func getTransactionIdFromHash(hash common.Hash) transaction.TransactionID {
	var transactionId transaction.TransactionID
	copy(transactionId[:], hash[:])
	return transactionId
}
