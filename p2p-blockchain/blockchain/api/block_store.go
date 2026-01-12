package api

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
)

type BlockStoreAPI interface {
	// AddBlock adds a new block to the block store.
	//
	// The block is linked to its parent based on the PreviousBlockHash field.
	// If the parent does not exist, the block becomes an orphan (added to roots without parent).
	//
	// After execution, the block store will contain the new block, either as part of the main chain, a side chain, or as an orphan.
	// This operation is idempotent; adding the same block multiple times has no effect after the first addition.
	//
	// Panics if the block violates any domain rules.
	// Returns a (non-nil) slice of block hashes that were added to the main or side chain because of this addition.
	AddBlock(block block.Block) (addedBlockHashes []common.Hash)

	// IsOrphanBlock checks if the given block is an orphan.
	// Returns an error if the block is not found in the block store.
	// O(1) time complexity.
	IsOrphanBlock(block block.Block) (bool, error)

	// IsPartOfMainChain checks if the given block is part of the main chain.
	// Returns false if the block is not found in the block store.
	// O(h) time complexity, with h being the height of the main chain.
	IsPartOfMainChain(block block.Block) bool

	// GetBlockByHash retrieves a block by its hash.
	// Returns an error if the block is not found.
	// O(1) time complexity.
	// Also works for orphan blocks.
	GetBlockByHash(hash common.Hash) (block.Block, error)

	// GetBlocksByHeight retrieves all blocks at the specified height.
	// Returns an empty slice if no blocks are found at that height.
	// Starts search at leaves, so retrieving blocks at higher heights is faster.
	// Never contains orphans, as orphans have no defined height in relation to the genesis block.
	GetBlocksByHeight(height uint64) []block.Block

	// GetCurrentHeight returns the current maximum height of the blockchain.
	// Note that this height may correspond to multiple(!) blocks in case of side chains.
	// Also note that the highest block may not be part of the main chain (the main chain is the one with the most accumulated work).
	// Use GetMainChainTip to get the tip of the main chain.
	GetCurrentHeight() uint64

	// GetMainChainTip returns the tip block of the main chain.
	// The main chain is defined as the chain with the highest accumulated work.
	// In case of multiple chains with the same accumulated work, one of them is returned arbitrarily(!).
	GetMainChainTip() block.Block
}
