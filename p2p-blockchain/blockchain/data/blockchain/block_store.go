// Package blockchain provides data structures and functions for managing the blockchain.
package blockchain

import "s3b/vsp-blockchain/p2p-blockchain/blockchain/data/block"

type BlockNode struct {
	// AccumulatedWork is the total chain work from genesis up to and including this header.
	AccumulatedWork uint64
	// Height is the number of blocks from genesis to this header (genesis has height 0).
	Height uint64

	Header *block.BlockHeader
	Block  *block.Block

	Parent   *BlockNode
	Children []*BlockNode
}

type BlockStore struct {
	HashToHeaders   map[block.Hash]*BlockNode
	HeightToHeaders map[uint64][]*BlockNode
}

func NewBlockStore() *BlockStore {
	return &BlockStore{
		HashToHeaders:   make(map[block.Hash]*BlockNode),
		HeightToHeaders: make(map[uint64][]*BlockNode),
	}
}

func (hs *BlockStore) AddBlock(block *block.Block) {

}
