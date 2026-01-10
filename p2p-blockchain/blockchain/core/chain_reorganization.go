package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/blockchain"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/utxo"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/utxopool"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
)

type ChainReorganization struct {
	lastKnownTip common.Hash
	blockStore   *blockchain.BlockStore
	utxoService  utxo.UTXOService
	mempool      *Mempool
}

func NewChainReorganization(
	blockStore *blockchain.BlockStore,
	utxoService utxo.UTXOService,
	mempool *Mempool,
) *ChainReorganization {
	return &ChainReorganization{
		blockStore:  blockStore,
		utxoService: utxoService,
		mempool:     mempool,
	}
}

// CheckAndReorganize checks if the new tip is different from the last known tip
// and triggers reorganization if necessary.
// Returns true if reorganization was performed.
func (cr *ChainReorganization) CheckAndReorganize(newTip common.Hash) (bool, error) {
	// First time initialization
	if cr.lastKnownTip == (common.Hash{}) {
		cr.updateLastKnownTip(newTip)
		return false, nil
	}

	// No change, no reorganization needed
	if cr.lastKnownTip == newTip {
		return false, nil
	}

	// Find the fork point
	forkPoint, err := cr.findForkPoint(cr.lastKnownTip, newTip)
	if err != nil {
		return false, err
	}

	// Phase 1: Disconnect old chain
	disconnectedBlocks, err := cr.disconnectBlocks(cr.lastKnownTip, forkPoint)
	if err != nil {
		return false, err
	}

	// Phase 2: Connect new chain
	newChainPath, err := cr.getNewChainPath(forkPoint, newTip)
	if err != nil {
		return false, err
	}

	err = cr.connectBlocks(newChainPath)
	if err != nil {
		return false, err
	}

	// Cleanup mempool - remove confirmed transactions and re-validate remaining ones
	allAffectedBlocks := append(disconnectedBlocks, cr.blockHashesFromBlocks(newChainPath)...)
	cr.cleanMempool(allAffectedBlocks)

	// Update the tip
	cr.updateLastKnownTip(newTip)

	return true, nil
}

// =========================================================================
// Phase 1: Disconnect (Rollback)
// =========================================================================

// findForkPoint finds the common ancestor (fork point) between the current tip
// and the new chain tip.
func (cr *ChainReorganization) findForkPoint(oldTip, newTip common.Hash) (common.Hash, error) {
	// Walk backwards from both tips until we find a common ancestor
	oldBlock, err := cr.blockStore.GetBlockByHash(oldTip)
	if err != nil {
		return common.Hash{}, err
	}

	newBlock, err := cr.blockStore.GetBlockByHash(newTip)
	if err != nil {
		return common.Hash{}, err
	}

	// Create sets to track visited blocks from both chains
	oldChain := make(map[common.Hash]bool)

	// Walk up the old chain and mark all ancestors
	currentHash := oldTip
	for {
		oldChain[currentHash] = true

		currentBlock, err := cr.blockStore.GetBlockByHash(currentHash)
		if err != nil {
			break
		}

		// Stop at genesis (no previous block)
		if currentBlock.Header.PreviousBlockHash == (common.Hash{}) {
			break
		}

		currentHash = currentBlock.Header.PreviousBlockHash
	}

	// Walk up the new chain until we find a block that's in the old chain
	currentHash = newTip
	for {
		if oldChain[currentHash] {
			// Found the fork point
			return currentHash, nil
		}

		currentBlock, err := cr.blockStore.GetBlockByHash(currentHash)
		if err != nil {
			break
		}

		// Stop at genesis
		if currentBlock.Header.PreviousBlockHash == (common.Hash{}) {
			// If we reached genesis and didn't find a common ancestor, genesis is the fork point
			return currentHash, nil
		}

		currentHash = currentBlock.Header.PreviousBlockHash
	}

	return common.Hash{}, err
}

// disconnectBlocks rolls back blocks from the current tip to the fork point.
// Returns the list of disconnected block hashes.
func (cr *ChainReorganization) disconnectBlocks(fromTip, toForkPoint common.Hash) ([]common.Hash, error) {
	disconnectedHashes := make([]common.Hash, 0)
	currentHash := fromTip

	// Walk backwards from tip to fork point
	for currentHash != toForkPoint {
		// Disconnect the current block
		err := cr.disconnectBlock(currentHash)
		if err != nil {
			return nil, err
		}

		disconnectedHashes = append(disconnectedHashes, currentHash)

		// Move to previous block
		currentBlock, err := cr.blockStore.GetBlockByHash(currentHash)
		if err != nil {
			return nil, err
		}

		currentHash = currentBlock.Header.PreviousBlockHash
	}

	return disconnectedHashes, nil
}

// disconnectBlock performs rollback operations for a single block:
// - Reverts UTXO state changes
// - Moves transactions back to mempool
func (cr *ChainReorganization) disconnectBlock(blockHash common.Hash) error {
	blk, err := cr.blockStore.GetBlockByHash(blockHash)
	if err != nil {
		return err
	}

	// First, undo the UTXO state changes
	err = cr.undoBlockState(blk)
	if err != nil {
		return err
	}

	// Then, move transactions back to mempool
	cr.moveTransactionsToMempool(blk)

	return nil
}

// undoBlockState reverts the UTXO state changes made by the given block.
// This includes:
// - Removing UTXOs created by the block's transactions
// - Restoring UTXOs that were spent by the block's transactions
func (cr *ChainReorganization) undoBlockState(blk block.Block) error {
	// Process transactions in reverse order
	for i := len(blk.Transactions) - 1; i >= 0; i-- {
		tx := blk.Transactions[i]
		txID := tx.TransactionId()

		// Save input UTXOs before reverting (needed for restoration)
		inputUTXOs, err := cr.saveInputUTXOs(&tx)
		if err != nil {
			// For coinbase transactions, there are no inputs to save
			inputUTXOs = []utxopool.UTXOEntry{}
		}

		// Revert the transaction (removes outputs, restores inputs)
		err = cr.utxoService.RevertTransaction(&tx, txID, inputUTXOs)
		if err != nil {
			return err
		}
	}

	return nil
}

// moveTransactionsToMempool moves all non-coinbase transactions from a block
// back to the mempool.
func (cr *ChainReorganization) moveTransactionsToMempool(blk block.Block) {
	for _, tx := range blk.Transactions {
		// Skip coinbase transactions (they are block-specific and cannot be in mempool)
		if tx.IsCoinbase() {
			continue
		}

		// Add transaction to mempool (if it's still valid, it will be accepted)
		cr.mempool.AddTransaction(tx)
	}
}

// =========================================================================
// Phase 2: Connect (Roll Forward)
// =========================================================================

// getNewChainPath returns the list of blocks from the fork point to the new tip.
// The blocks are returned in the order they should be applied (fork point + 1 to new tip).
func (cr *ChainReorganization) getNewChainPath(forkPoint, newTip common.Hash) ([]block.Block, error) {
	path := make([]block.Block, 0)

	// Walk backwards from new tip to fork point, collecting blocks
	currentHash := newTip
	for currentHash != forkPoint {
		blk, err := cr.blockStore.GetBlockByHash(currentHash)
		if err != nil {
			return nil, err
		}

		// Prepend to path (so we end up with fork point+1 -> new tip order)
		path = append([]block.Block{blk}, path...)

		currentHash = blk.Header.PreviousBlockHash
	}

	return path, nil
}

// connectBlocks applies blocks from the new chain path.
func (cr *ChainReorganization) connectBlocks(blocks []block.Block) error {
	for _, blk := range blocks {
		err := cr.connectBlock(blk)
		if err != nil {
			return err
		}
	}
	return nil
}

// connectBlock performs operations to apply a single block:
// - Applies transactions (updates UTXO set)
// - Removes confirmed transactions from mempool
func (cr *ChainReorganization) connectBlock(blk block.Block) error {
	// Apply the block's transactions to the UTXO set
	err := cr.applyBlock(blk)
	if err != nil {
		return err
	}

	// Clean mempool of confirmed transactions
	cr.cleanMempool([]common.Hash{blk.Hash()})

	return nil
}

// applyBlock applies all transactions in the block to the UTXO set.
func (cr *ChainReorganization) applyBlock(blk block.Block) error {
	// Get block height from blockStore
	// We need this to properly mark UTXOs as confirmed
	blockHeight := cr.getBlockHeight(blk)

	for _, tx := range blk.Transactions {
		txID := tx.TransactionId()
		isCoinbase := tx.IsCoinbase()

		// Apply the transaction to the UTXO set
		err := cr.utxoService.ApplyTransaction(&tx, txID, blockHeight, isCoinbase)
		if err != nil {
			return err
		}
	}

	return nil
}

// cleanMempool removes confirmed transactions from the mempool.
// Also re-validates remaining transactions to remove conflicts.
func (cr *ChainReorganization) cleanMempool(blockHashes []common.Hash) {
	// Use the existing mempool Remove method which handles:
	// 1. Removing confirmed transactions
	// 2. Removing transactions that conflict with confirmed ones
	// 3. Re-validating remaining transactions
	cr.mempool.Remove(blockHashes)
}

// =========================================================================
// Helper methods
// =========================================================================

// saveInputUTXOs retrieves and saves the input UTXOs of a transaction
// before they are spent. This is needed for later rollback.
func (cr *ChainReorganization) saveInputUTXOs(tx *transaction.Transaction) ([]utxopool.UTXOEntry, error) {
	inputUTXOs := make([]utxopool.UTXOEntry, 0, len(tx.Inputs))

	for _, input := range tx.Inputs {
		outpoint := utxopool.NewOutpoint(input.PrevTxID, input.OutputIndex)

		// Retrieve the UTXO entry from the UTXO set
		entry, err := cr.utxoService.GetUTXOEntry(outpoint)
		if err != nil {
			return nil, err
		}

		inputUTXOs = append(inputUTXOs, entry)
	}

	return inputUTXOs, nil
}

// blockHashesFromBlocks extracts block hashes from a list of blocks.
func (cr *ChainReorganization) blockHashesFromBlocks(blocks []block.Block) []common.Hash {
	hashes := make([]common.Hash, len(blocks))
	for i, blk := range blocks {
		hashes[i] = blk.Hash()
	}
	return hashes
}

// getBlockHeight retrieves the height of a block from the blockStore.
func (cr *ChainReorganization) getBlockHeight(blk block.Block) uint64 {
	// Walk backwards to genesis counting blocks
	// This is a simple implementation - could be optimized by caching heights
	height := uint64(0)
	currentHash := blk.Hash()

	for {
		currentBlock, err := cr.blockStore.GetBlockByHash(currentHash)
		if err != nil {
			break
		}

		// Stop at genesis (no previous block)
		if currentBlock.Header.PreviousBlockHash == (common.Hash{}) {
			break
		}

		height++
		currentHash = currentBlock.Header.PreviousBlockHash
	}

	return height
}

// updateLastKnownTip updates the last known chain tip.
func (cr *ChainReorganization) updateLastKnownTip(newTip common.Hash) {
	cr.lastKnownTip = newTip
}
