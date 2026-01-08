// Package blockchain provides data structures and functions for managing the blockchain.
package blockchain

import (
	"fmt"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"slices"

	"bjoernblessin.de/go-utils/util/assert"
	mapset "github.com/deckarep/golang-set/v2"
)

// blockForest represents a collection of trees structures representing the blockchain.
//
// There is one main chain (starting with the genesis as root), potentially several side chains and orphans.
//   - The main chain is the chain with the highest accumulated work. This also means that the main chain does not have to be the longest chain but typically is.
//   - Side chains are chains that branch off from the main chain at some point.
//   - Orphans are blocks that don't have the genesis block as an ancestor. Their highest ancestor is a root blockNode in the forest which is not the genesis block.
type blockForest struct {
	// Roots are the root nodes of all trees in the forest.
	// The genesis block is always one of the roots.
	// Also includes orphans.
	Roots []*blockNode
	// Leaves are the leaf nodes of all trees in the forest.
	// This includes the tip of the main chain and side chains and orphan blocks.
	Leaves []*blockNode
}

// blockNode represents a node in the block forest.
type blockNode struct {
	// AccumulatedWork is the total chain work from genesis up to and including this header.
	// There is no accumulated work for orphans.
	AccumulatedWork uint64
	// Height is the number of blocks from genesis to this header (genesis has height 0).
	Height uint64

	Block *block.Block

	Parent   *blockNode
	Children []*blockNode
}

// BlockStore manages the storage and retrieval of blocks in the blockchain.
// Its underlying structure is private and only [block.Block] can be accessed externally.
// It's responsible for maintaining the integrity of the blockchain data. This means the BlockStore state is always consistent and valid.
// Retrieved blocks are always copies of the stored blocks to prevent external modification of the internal state. This means that any modification to a retrieved block does not affect the stored block in the BlockStore.
type BlockStore struct {
	blockForest blockForest
	// hashToHeaders provides fast lookup of blockNodes by their hash.
	hashToHeaders map[common.Hash]*blockNode
}

func NewBlockStore(genesis block.Block) *BlockStore {
	genesisNode := blockNode{
		AccumulatedWork: uint64(genesis.BlockDifficulty()),
		Height:          0,
		Block:           &genesis,
		Parent:          nil,
		Children:        []*blockNode{},
	}

	return &BlockStore{
		blockForest: blockForest{
			Roots:  []*blockNode{&genesisNode},
			Leaves: []*blockNode{&genesisNode},
		},
		hashToHeaders: map[common.Hash]*blockNode{
			genesis.Hash(): &genesisNode,
		},
	}
}

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
func (s *BlockStore) AddBlock(block block.Block) (addedBlockHashes []common.Hash) {
	blockHash := block.Hash()

	// TODO validate block, panic if invalid

	addedBlockHashes = []common.Hash{}

	// Check if block already exists
	if _, exists := s.hashToHeaders[blockHash]; exists {
		return
	}

	// Find parent
	parentHash := block.Header.PreviousBlockHash
	parent, parentExists := s.hashToHeaders[parentHash]
	isParentOrphan := false
	if parentExists {
		var err error
		isParentOrphan, err = s.IsOrphanBlock(*parent.Block)
		assert.IsNil(err, "error checking if parent is orphan")
	}

	newNode := blockNode{
		Block:    &block,
		Children: []*blockNode{},
	}

	// Add new block to hash map
	s.hashToHeaders[blockHash] = &newNode

	if parentExists && !isParentOrphan {
		// Block can be part of main chain or side chain
		s.connectNodes(parent, &newNode)
		addedBlockHashes = append(addedBlockHashes, blockHash)

		addedBlockHashes = append(addedBlockHashes, s.connectOrphanBlock(&newNode)...)
	} else {
		// Block is an orphan
		s.blockForest.Roots = append(s.blockForest.Roots, &newNode)
	}

	return
}

// connectOrphanBlock tries to connect orphan blocks to the given newNode.
// Returns a slice of block hashes that were added to the main or side chain because of this connection.
func (s *BlockStore) connectOrphanBlock(newNode *blockNode) (addedBlockHashes []common.Hash) {
	addedBlockHashes = []common.Hash{}

	// Collect orphans to connect first to avoid modifying slice during iteration (connectNodes modifies s.blockForest.Roots)
	var orphansToConnect []*blockNode
	for _, orphanRoot := range s.blockForest.Roots {
		if orphanRoot.Block.Header.PreviousBlockHash == newNode.Block.Hash() {
			orphansToConnect = append(orphansToConnect, orphanRoot)
		}
	}

	for _, orphanRoot := range orphansToConnect {
		// Connects (newNode) --> (orphanRoot)
		s.connectNodes(newNode, orphanRoot)
		addedBlockHashes = append(addedBlockHashes, orphanRoot.Block.Hash())

		// Recursively try to connect further orphans
		addedBlockHashes = append(addedBlockHashes, s.connectOrphanBlock(orphanRoot)...)
	}

	return
}

// connectNodes connects a parent blockNode to a child blockNode.
// Updates (1) accumulated work, (2) height, (3) leaves, (4) roots and (5) connection relation accordingly.
func (s *BlockStore) connectNodes(parent *blockNode, child *blockNode) {
	child.Parent = parent
	parent.Children = append(parent.Children, child)

	child.AccumulatedWork = parent.AccumulatedWork + uint64(child.Block.BlockDifficulty())
	child.Height = parent.Height + 1

	// Remove parent from leaves (if it was a leaf)
	// Is a leaf if it was a tip before
	// Is not a leaf if it had children before (creating a side chain)
	s.blockForest.Leaves = slices.DeleteFunc(s.blockForest.Leaves, func(leaf *blockNode) bool {
		return leaf == parent
	})

	// Add child to leaves
	s.blockForest.Leaves = append(s.blockForest.Leaves, child)

	// Remove child from roots (if it was a root)
	// Is a root if it was an orphan before
	// Is not a root if completely new block
	s.blockForest.Roots = slices.DeleteFunc(s.blockForest.Roots, func(root *blockNode) bool {
		return root == child
	})
}

// IsOrphanBlock checks if the given block is an orphan.
// Returns an error if the block is not found in the block store.
// O(1) time complexity.
func (s *BlockStore) IsOrphanBlock(block block.Block) (bool, error) {
	blockNode, exists := s.hashToHeaders[block.Hash()]
	if !exists {
		return false, fmt.Errorf("block with hash %v not found", block.Hash())
	}

	// A block is an orphan if it has no parent and no accumulated work (genesis has accumulated work from its difficulty)
	isOrphan := blockNode.Parent == nil && blockNode.AccumulatedWork == 0
	return isOrphan, nil
}

// IsPartOfMainChain checks if the given block is part of the main chain.
// Returns false if the block is not found in the block store.
// O(h) time complexity, with h being the height of the main chain.
func (s *BlockStore) IsPartOfMainChain(block block.Block) bool {
	blockNode, exists := s.hashToHeaders[block.Hash()]
	if !exists {
		return false
	}

	// Get main chain tip
	mainChainTip := s.GetMainChainTip()
	mainChainTipNode := s.hashToHeaders[mainChainTip.Hash()]

	// Traverse up from main chain tip to genesis, checking if we encounter the blockNode
	currentNode := mainChainTipNode
	for currentNode != nil {
		if currentNode == blockNode {
			return true
		}
		currentNode = currentNode.Parent
	}

	return false
}

// GetBlockByHash retrieves a block by its hash.
// Returns an error if the block is not found.
// O(1) time complexity.
// Also works for orphan blocks.
func (s *BlockStore) GetBlockByHash(hash common.Hash) (block.Block, error) {
	blockNode, exists := s.hashToHeaders[hash]
	if !exists {
		return block.Block{}, fmt.Errorf("block with hash %v not found", hash)
	}

	return copyOfBlock(blockNode), nil
}

// GetBlocksbyHeight retrieves all blocks at the specified height.
// Returns an empty slice if no blocks are found at that height.
// Starts search at leaves, so retrieving blocks at higher heights is faster.
// Never contains orphans, as orphans have no defined height in relation to the genesis block.
func (s *BlockStore) GetBlocksByHeight(height uint64) []block.Block {
	blockHashesAtHeight := mapset.NewSet[common.Hash]()

	for _, leaf := range s.blockForest.Leaves {
		currentNode := leaf

		// Traverse up the tree as long as (1) we are not at the root and (2) we are above the desired height
		for currentNode != nil && currentNode.Height >= height {
			if currentNode.Height == height {
				// isOrphan removes orphans from the result (orphans may have the default height 0)
				isOrphan, err := s.IsOrphanBlock(*currentNode.Block)
				assert.IsNil(err, "error checking if block is orphan")
				if !isOrphan {
					blockHashesAtHeight.Add(currentNode.Block.Hash())
				}

				break
			}

			currentNode = currentNode.Parent
		}
	}

	// Convert hashes back to blocks
	blocksAtHeight := make([]block.Block, 0, blockHashesAtHeight.Cardinality())
	for hash := range blockHashesAtHeight.Iter() {
		b, err := s.GetBlockByHash(hash)
		assert.IsNil(err, "error retrieving block by hash")
		blocksAtHeight = append(blocksAtHeight, b)
	}

	return blocksAtHeight
}

// GetCurrentHeight returns the current maximum height of the blockchain.
// Note that this height may correspond to multiple(!) blocks in case of side chains.
// Also note that the highest block may not be part of the main chain (the main chain is the one with the most accumulated work).
// Use GetMainChainTip to get the tip of the main chain.
func (s *BlockStore) GetCurrentHeight() uint64 {
	var maxHeight uint64
	for _, leaf := range s.blockForest.Leaves {
		if leaf.Height > maxHeight {
			maxHeight = leaf.Height
		}
	}

	return maxHeight
}

// GetMainChainTip returns the tip block of the main chain.
// The main chain is defined as the chain with the highest accumulated work.
// In case of multiple chains with the same accumulated work, one of them is returned arbitrarily(!).
func (s *BlockStore) GetMainChainTip() block.Block {
	var mainChainTip *blockNode
	var maxAccumulatedWork uint64

	for _, leaf := range s.blockForest.Leaves {
		if leaf.AccumulatedWork > maxAccumulatedWork {
			maxAccumulatedWork = leaf.AccumulatedWork
			mainChainTip = leaf
		}
	}

	assert.IsNotNil(mainChainTip, "no main chain tip found in block store")

	return copyOfBlock(mainChainTip)
}

// copyOfBlock creates and returns a deep copy of the block contained in the given blockNode.
// Should be used to prevent external modification of the internal state of the BlockStore.
func copyOfBlock(node *blockNode) block.Block {
	originalBlock := node.Block

	// Deep copy transactions
	copiedTransactions := make([]transaction.Transaction, len(originalBlock.Transactions))
	copy(copiedTransactions, originalBlock.Transactions)

	// Create new block with copied transactions
	copiedBlock := block.Block{
		Header:       originalBlock.Header,
		Transactions: copiedTransactions,
	}

	return copiedBlock
}
