package miner

import (
	"math/big"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
)

type MinerAPI interface {
	CreateCandidateBlock() block.Block
	MineBlock(candidateBlock block.Block) block.Block
}

type MinerService struct {
}

func (m *MinerService) CreateCandidateBlock(transactions []transaction.Transaction) block.Block {
	header := createCandidateBlockHeader()

	return block.Block{Header: header}
}

// TODO: Kapitel Die Block-Header aufbauen
func createCandidateBlockHeader() block.BlockHeader {
	return block.BlockHeader{}
}

// MineBlock Mines a block by change the nonce until the block matches the given difficulty target
func (m *MinerService) MineBlock(candidateBlock block.Block) (nonce uint32) {
	target := getTarget(candidateBlock.Header.DifficultyTarget)

	var hashInt big.Int
	nonce = 0

	for {
		candidateBlock.Header.Nonce = nonce
		hash := candidateBlock.Hash() //TODO: Will work after merging https://github.com/bjoern621/VSP-Blockchain/pull/174

		hashInt.SetBytes(hash[:])
		if hashInt.Cmp(&target) == -1 {
			break
		}
		nonce++
	}

	return nonce
}

// getTarget calculates the target for the proof of work algorithm
// It does so by shifting a one in a 256 bit number to the left by 256 - difficultyBits.
// Theory: 0b1 << (256 - difficultyBits) But this is not possible as Go has no operator overloading :( and so big.Int is used
// This is required as a valid hash should be smaller than the target.
func getTarget(difficulty uint32) big.Int {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-difficulty))

	return *target
}
