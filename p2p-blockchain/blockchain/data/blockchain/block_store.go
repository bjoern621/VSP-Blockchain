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

// Wahrscheinlich wird hier dann bei dir intern entschieden, an welche Kette, oder sogar orphan pool, der block angehängt wird
// func (hs *BlockStore) AddBlock(block *block.Block) (bool, error) {

// }

// func RecheckOrphanBlocks() (blocksAcceptedIntoChain int) {}

// // IsOrphanBlock return true if the given block is an orphan block.
// // An orphan block is a block that is not connected to the main chain or any known side chain.
// func IsOrphanBlock(block block.Block) bool                 {} // Wird zB. gebraucht, um ein GetHeaders/GetData zu starten, falls true zurück kommt
// func IsPartOfMainChain(receivedBlock block.Block) bool     {} //Wird gebraucht, da abhängig davon, ob der block Teil der MainChain ist, oder nicht, anders verfahrenn wird
// func GetBlockByHash(hash common.Hash) (block.Block, error) {}
// func GetBlockByIndex(height uint32) (block.Block, error)   {}
// func GetCurrentHeight() uint32
