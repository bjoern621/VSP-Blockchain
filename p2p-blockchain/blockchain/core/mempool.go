package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/validation"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/blockchain"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"

	"sync"

	"bjoernblessin.de/go-utils/util/assert"
	"bjoernblessin.de/go-utils/util/logger"
)

type Mempool struct {
	validator  validation.TransactionValidatorAPI
	blockStore blockchain.BlockStoreAPI

	transactions map[transaction.TransactionID]transaction.Transaction
	lock         sync.Mutex
}

func NewMempool(validator validation.TransactionValidatorAPI, blockStore blockchain.BlockStoreAPI) *Mempool {
	return &Mempool{
		validator:    validator,
		blockStore:   blockStore,
		transactions: make(map[transaction.TransactionID]transaction.Transaction),
	}
}

// GetTransactionsForMining returns all transactions that are eligible for mining
func (m *Mempool) GetTransactionsForMining() []transaction.Transaction {
	m.lock.Lock()
	defer m.lock.Unlock()

	txs := m.getTransactionsForMining()

	logger.Infof("[mempool] Returning %d transaction eligible for mining", len(txs))
	return txs
}

func (m *Mempool) getTransactionsForMining() []transaction.Transaction {
	txs := make([]transaction.Transaction, 0, len(m.transactions))
	for _, tx := range m.transactions {
		txs = append(txs, tx)
	}

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
	mainChainTip := m.blockStore.GetMainChainTip()
	mainChainTipHash := mainChainTip.Hash()
	ok, err := m.validator.ValidateTransaction(tx, mainChainTipHash)
	assert.Assert(ok && err == nil, "Transaction is invalid: %v", err)

	m.lock.Lock()
	defer m.lock.Unlock()

	txId := tx.TransactionId()
	_, ok = m.transactions[txId]
	if !ok {
		logger.Infof("[mempool] Adding new transaction %v to the mempool", txId)
		m.transactions[txId] = tx
	} else {
		logger.Infof("[mempool] Transaction %v is already known to the mempool", txId)
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

	for txId := range m.transactions {
		// Skip if already marked for removal
		if toRemove[txId] {
			continue
		}

		tx := m.transactions[txId]

		mainChainTip := m.blockStore.GetMainChainTip()
		mainChainTipHash := mainChainTip.Hash()
		ok, _ := m.validator.ValidateTransaction(tx, mainChainTipHash)
		if !ok {
			// Transaction is no longer valid (UTXOs spent, conflicts with confirmed tx)
			toRemove[txId] = true
		}
	}

	removeCount := 0
	// Remove all marked transactions
	for txId := range toRemove {
		_, exists := m.transactions[txId]
		if exists {
			removeCount++
		}
		delete(m.transactions, txId)
	}

	logger.Infof("[mempool] Removed %d transactions from the mempool after new block arrived", removeCount)
}

// GetAllTransactionHashes returns all transaction hashes currently in the mempool.
func (m *Mempool) GetAllTransactionHashes() []common.Hash {
	m.lock.Lock()
	defer m.lock.Unlock()

	hashes := make([]common.Hash, 0, len(m.transactions))
	for txId := range m.transactions {
		var hash common.Hash
		copy(hash[:], txId[:])
		hashes = append(hashes, hash)
	}
	return hashes
}

func getTransactionIdFromHash(hash common.Hash) transaction.TransactionID {
	var transactionId transaction.TransactionID
	copy(transactionId[:], hash[:])
	return transactionId
}

// GetTransactionByHash retrieves the transaction with the given hash from the mempool.
// Returns the transaction and true if found, or false if not found.
func (m *Mempool) GetTransactionByHash(hash common.Hash) (transaction.Transaction, bool) {
	txId := getTransactionIdFromHash(hash)
	m.lock.Lock()
	defer m.lock.Unlock()

	tx, ok := m.transactions[txId]
	return tx, ok
}
