package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/blockchain"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/validation"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"

	"sync"

	"bjoernblessin.de/go-utils/util/assert"
)

type Mempool struct {
	validator  validation.ValidationAPI
	blockStore blockchain.BlockStoreAPI

	transactions map[transaction.TransactionID]transaction.Transaction
	lock         sync.Mutex
}

func NewMempool(validator validation.ValidationAPI, blockStore blockchain.BlockStoreAPI) *Mempool {
	return &Mempool{
		validator:    validator,
		blockStore:   blockStore,
		transactions: make(map[transaction.TransactionID]transaction.Transaction),
	}
}

func (m *Mempool) GetTransactionsForMining() []transaction.Transaction {
	m.lock.Lock()
	defer m.lock.Unlock()

	txs := make([]transaction.Transaction, 0, len(m.transactions))
	for _, tx := range m.transactions {
		txs = append(txs, tx)
	}

	clear(m.transactions)
	return txs
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

	// Collect all transactions from the given blocks
	for _, blockHash := range blockHashes {
		blk, err := m.blockStore.GetBlockByHash(blockHash)
		if err == nil {
			// Mark all transactions in this block for removal
			for _, tx := range blk.Transactions {
				txId := tx.TransactionId()
				toRemove[txId] = true
			}
		}
	}

	// Re-validate each transaction and remove invalid ones
	for txId := range m.transactions {
		// Skip if already marked for removal
		if toRemove[txId] {
			continue
		}

		// Re-validate each transaction
		tx := m.transactions[txId]
		ok, _ := m.validator.ValidateTransaction(&tx)
		if !ok {
			// Transaction is no longer valid (UTXOs spent, conflicts with confirmed tx)
			toRemove[txId] = true
		}
	}

	// Remove all marked transactions
	for txId := range toRemove {
		delete(m.transactions, txId)
	}
}

func getTransactionIdFromHash(hash common.Hash) transaction.TransactionID {
	var transactionId transaction.TransactionID
	copy(transactionId[:], hash[:])
	return transactionId
}
