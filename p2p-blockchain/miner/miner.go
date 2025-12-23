package miner

import (
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

func createCandidateBlockHeader() block.BlockHeader {
	return block.BlockHeader{}
}

func (m *MinerService) MineBlock(candidateBlock block.Block) block.Block {

	return block.Block{}
}
