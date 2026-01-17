package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/utxo"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/blockchain"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
)

// ChainReorganizationAPI defines the interface for chain reorganization handling.
// This is the only place where the tx-mempool is directly manipulated during reorgs.
type ChainReorganizationAPI interface {
	// CheckAndReorganize checks if the new tip is different from the last known tip
	// and triggers reorganization if necessary.
	// Returns true if a reorganization was performed (i.e., part of the old chain was disconnected).
	// Returns false if no change occurred or if the new blocks are a simple extension of the current chain.
	// This includes updating the UTXO set and mempool accordingly in both cases.
	// This is usually called after one or more blocks are added to the block store.
	CheckAndReorganize(newTip common.Hash) (bool, error)
}

type ChainReorganization struct {
	lastKnownTip      common.Hash
	blockStore        blockchain.BlockStoreAPI
	multiChainService *utxo.MultiChainUTXOService
	mempool           *Mempool
}

func NewChainReorganization(
	blockStore blockchain.BlockStoreAPI,
	multiChainService *utxo.MultiChainUTXOService,
	mempool *Mempool,
) *ChainReorganization {
	return &ChainReorganization{
		blockStore:        blockStore,
		multiChainService: multiChainService,
		mempool:           mempool,
	}
}

// CheckAndReorganize checks if the new tip is different from the last known tip
// and triggers reorganization if necessary.
//
// This method uses MultiChainUTXOService to:
// - Promote side chains to main chain (with proper delta handling)
// - Preserve old main chain as a side chain (for potential future reorgs)
// - Clear and rebuild mempool after reorg
//
// Returns true if a reorganization was performed (i.e., part of the old chain was disconnected).
// Returns false if no change occurred or if the new blocks are a simple extension of the current chain.
func (cr *ChainReorganization) CheckAndReorganize(newTip common.Hash) (bool, error) {
	// First time initialization
	if cr.lastKnownTip == (common.Hash{}) {
		cr.updateLastKnownTip(newTip)
		cr.multiChainService.SetMainChainTip(newTip)
		return false, nil
	}

	// No change, no reorganization needed
	if cr.lastKnownTip == newTip {
		return false, nil
	}

	// Check if newTip is a direct extension of the current chain
	newBlock, err := cr.blockStore.GetBlockByHash(newTip)
	if err != nil {
		return false, err
	}

	// Case 1: Simple chain extension - NOT a reorganization
	// The new block(s) build directly on top of our current tip
	// NOTE: Block handler already applied the UTXO changes and updated the tip
	if newBlock.Header.PreviousBlockHash == cr.lastKnownTip {
		cr.updateLastKnownTip(newTip)
		return false, nil // NO reorg occurred, just chain extension
	}

	// Case 2: Reorganization needed - the new chain diverges from our current chain
	// Find the fork point
	forkPoint, err := cr.findForkPoint(cr.lastKnownTip, newTip)
	if err != nil {
		return false, err
	}

	// Collect blocks to disconnect and connect
	disconnectedBlocks, err := cr.collectBlocksToDisconnect(cr.lastKnownTip, forkPoint)
	if err != nil {
		return false, err
	}

	connectedBlocks, err := cr.collectBlocksToConnect(forkPoint, newTip)
	if err != nil {
		return false, err
	}

	// Use MultiChainUTXOService to perform the reorganization
	err = cr.multiChainService.PromoteSideChainToMain(
		newTip,
		forkPoint,
		disconnectedBlocks,
		connectedBlocks,
	)
	if err != nil {
		return false, err
	}

	// Clear mempool - after reorg, all unconfirmed txs need re-validation
	// Transactions from disconnected blocks that are still valid will be re-relayed
	cr.multiChainService.ClearMempool()

	// Move transactions from disconnected blocks back to mempool
	for _, blk := range disconnectedBlocks {
		cr.moveTransactionsToMempool(blk)
	}

	// Remove transactions that are now confirmed in the new chain
	for _, blk := range connectedBlocks {
		cr.removeConfirmedFromMempool(blk)
	}

	// Update the tip
	cr.updateLastKnownTip(newTip)

	return true, nil // Reorg occurred
}

// =========================================================================
// Block Collection
// =========================================================================

// collectBlocksToDisconnect collects blocks from tip to fork point (exclusive).
// Returns blocks in reverse order (tip first, fork point not included).
func (cr *ChainReorganization) collectBlocksToDisconnect(fromTip, toForkPoint common.Hash) ([]block.Block, error) {
	blocks := make([]block.Block, 0)
	currentHash := fromTip

	for currentHash != toForkPoint {
		blk, err := cr.blockStore.GetBlockByHash(currentHash)
		if err != nil {
			return nil, err
		}

		blocks = append(blocks, blk)
		currentHash = blk.Header.PreviousBlockHash
	}

	return blocks, nil
}

// collectBlocksToConnect collects blocks from fork point to new tip.
// Returns blocks in forward order (fork point + 1 first, new tip last).
func (cr *ChainReorganization) collectBlocksToConnect(forkPoint, newTip common.Hash) ([]block.Block, error) {
	blocks := make([]block.Block, 0)
	currentHash := newTip

	for currentHash != forkPoint {
		blk, err := cr.blockStore.GetBlockByHash(currentHash)
		if err != nil {
			return nil, err
		}

		// Prepend to get forward order
		blocks = append([]block.Block{blk}, blocks...)
		currentHash = blk.Header.PreviousBlockHash
	}

	return blocks, nil
}

// =========================================================================
// Fork Point Detection
// =========================================================================

// findForkPoint finds the common ancestor (fork point) between the current tip
// and the new chain tip.
func (cr *ChainReorganization) findForkPoint(oldTip, newTip common.Hash) (common.Hash, error) {
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
			return common.Hash{}, err
		}

		// Stop at genesis
		if currentBlock.Header.PreviousBlockHash == (common.Hash{}) {
			// If we reached genesis and didn't find a common ancestor, genesis is the fork point
			return currentHash, nil
		}

		currentHash = currentBlock.Header.PreviousBlockHash
	}
}

// =========================================================================
// Mempool Management
// =========================================================================

// moveTransactionsToMempool moves all non-coinbase transactions from a block
// back to the mempool for re-validation.
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

// removeConfirmedFromMempool removes transactions that are confirmed in a block.
func (cr *ChainReorganization) removeConfirmedFromMempool(blk block.Block) {
	for _, tx := range blk.Transactions {
		if !tx.IsCoinbase() {
			cr.mempool.RemoveTransaction(tx.TransactionId())
		}
	}
}

// =========================================================================
// Helper Methods
// =========================================================================

// updateLastKnownTip updates the last known chain tip.
func (cr *ChainReorganization) updateLastKnownTip(newTip common.Hash) {
	cr.lastKnownTip = newTip
}
