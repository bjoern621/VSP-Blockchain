package miner

import (
	"context"
	"fmt"
	"math/big"
	blockchainApi "s3b/vsp-blockchain/p2p-blockchain/blockchain/api"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api"

	"bjoernblessin.de/go-utils/util/logger"
)

type MinerAPI interface {
	StartMining(transactions []transaction.Transaction) (minedBlock block.Block)
	StopMining()
}

type MinerService struct {
	cancelMining        context.CancelFunc
	blockchainMsgSender api.BlockchainAPI
	blockchain          blockchainApi.BlockchainAPI
}

func (m *MinerService) StartMining(transactions []transaction.Transaction) {
	candidateBlock := m.createCandidateBlock(transactions)

	ctx, cancel := context.WithCancel(context.Background())
	m.cancelMining = cancel

	go func() {
		nonce, timestamp, err := m.mineBlock(candidateBlock, ctx)
		if err != nil {
			logger.Infof("Mining stopped: %v", err)
			return
		}
		candidateBlock.Header.Nonce = nonce
		candidateBlock.Header.Timestamp = timestamp
		logger.Infof("Mined new block: %v", candidateBlock.Header)
		m.blockchain.Block(candidateBlock, "")
	}()
}

func (m *MinerService) StopMining() {
	m.cancelMining()
}

func (m *MinerService) createCandidateBlock(transactions []transaction.Transaction) block.Block {
	header := createCandidateBlockHeader()

	return block.Block{Header: header}
}

// TODO: Kapitel Die Block-Header aufbauen
func createCandidateBlockHeader() block.BlockHeader {
	return block.BlockHeader{}
}

// MineBlock Mines a block by change the nonce until the block matches the given difficulty target
func (m *MinerService) mineBlock(candidateBlock block.Block, ctx context.Context) (nonce uint32, timestamp int64, err error) {
	target := getTarget(candidateBlock.Header.DifficultyTarget)

	var counter uint64 = 0
	var hashInt big.Int
	nonce = 0

	for {
		select {
		case <-ctx.Done():
			logger.Infof("Mining cancelled")
			return 0, 0, fmt.Errorf("mining cancelled")
		default:
			if counter > uint64(^uint32(0)) {
				timestamp++
				candidateBlock.Header.Timestamp = timestamp
				counter = 0
			}

			candidateBlock.Header.Nonce = nonce
			hash := candidateBlock.Hash()

			hashInt.SetBytes(hash[:])
			if hashInt.Cmp(&target) == -1 {
				return nonce, timestamp, nil
			}
			nonce++
			counter++
		}
	}
}

// getTarget calculates the target for the proof of work algorithm
// It does so by shifting a one in a 256 bit number to the left by 256 - difficultyBits.
// Theory: 0b1 << (256 - difficultyBits) But this is not possible as Go has no operator overloading :( and so big.Int is used
// This is required as a valid hash should be smaller than the target.
func getTarget(difficulty uint8) big.Int {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-uint32(difficulty)))

	return *target
}
