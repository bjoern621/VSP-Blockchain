// Package blockchain provides data structures and functions for managing the blockchain.
package blockchain

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
)

// blockForest represents a collection of trees structures representing the blockchain.
//
// There is one main chain (starting with the genesis as root), potentially several side chains and orphans.
//   - The main chain is the chain with the highest accumulated work. This also means that the main chain does not have to be the longest chain but typically is.
//   - Side chains are chains that branch off from the main chain at some point.
//   - Orphans are blocks that don't have the genesis block as an ancestor. Their highest ancestor is a root blockNode in the forest which is not the genesis block.
type blockForest struct {
	Roots  []*blockNode
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
// It's underlying structure is private and only [block.Block] can be accessed externally.
// It is responsible for maintaining the integrity of the blockchain data. This means the BlockStore state is always consistent and valid.
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
// Errors if the block violates any domain rules.
func (s *BlockStore) AddBlock(block block.Block) error {
	blockHash := block.Hash()

	// Check if block already exists
	if _, exists := s.hashToHeaders[blockHash]; exists {
		return nil // Block already exists, no need to add it again
	}

	// Find parent block
	parentHash := block.Header.PreviousBlockHash
	parent, parentExists := s.hashToHeaders[parentHash]

	newNode := blockNode{
		Block:    &block,
		Children: []*blockNode{},
	}

	if parentExists {
		// Block has a parent
		newNode.Parent = parent
		newNode.AccumulatedWork = parent.AccumulatedWork + uint64(block.BlockDifficulty())
		newNode.Height = parent.Height + 1

		// Add new node to parent's children
		parent.Children = append(parent.Children, &newNode)

		// Remove parent from leaves if it was a leaf
		for i, leaf := range s.blockForest.Leaves {
			if leaf == parent {
				s.blockForest.Leaves = append(s.blockForest.Leaves[:i], s.blockForest.Leaves[i+1:]...)
				break
			}
		}
	} else {
		// Block is an orphan (parent not found)
		newNode.AccumulatedWork = uint64(block.BlockDifficulty())
		newNode.Height = 0
		s.blockForest.Roots = append(s.blockForest.Roots, &newNode)
	}

	// Add new block to leaves and hash map
	s.blockForest.Leaves = append(s.blockForest.Leaves, &newNode)
	s.hashToHeaders[blockHash] = &newNode

	return nil
}

// func RecheckOrphanBlocks() (blocksAcceptedIntoChain int) {}

// // IsOrphanBlock return true if the given block is an orphan block.
// // An orphan block is a block that is not connected to the main chain or any known side chain.
// func IsOrphanBlock(block block.Block) bool                 {} // Wird zB. gebraucht, um ein GetHeaders/GetData zu starten, falls true zurück kommt
// func IsPartOfMainChain(receivedBlock block.Block) bool     {} //Wird gebraucht, da abhängig davon, ob der block Teil der MainChain ist, oder nicht, anders verfahrenn wird
// func GetBlockByHash(hash common.Hash) (block.Block, error) {}
// func GetBlockByIndex(height uint32) (block.Block, error)   {}
// func GetCurrentHeight() uint32
