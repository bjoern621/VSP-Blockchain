package core

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	blockchainApi "s3b/vsp-blockchain/p2p-blockchain/blockchain/api"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"s3b/vsp-blockchain/p2p-blockchain/miner/data"
	"sync"

	"bjoernblessin.de/go-utils/util/logger"
)

type minerService struct {
	mu            sync.RWMutex
	miningEnabled bool
	cancelMining  context.CancelFunc
	blockchain    blockchainApi.BlockchainAPI
	utxoService   blockchainApi.UtxoStoreAPI
	blockStore    blockchainApi.BlockStoreAPI

	ownPubKeyHash transaction.PubKeyHash
}

func NewMinerService(
	blockchain blockchainApi.BlockchainAPI,
	utxoServiceAPI blockchainApi.UtxoStoreAPI,
	blockStore blockchainApi.BlockStoreAPI,
) *minerService {
	return &minerService{
		blockchain:    blockchain,
		utxoService:   utxoServiceAPI,
		blockStore:    blockStore,
		miningEnabled: true,
		ownPubKeyHash: ChoosePubKeyHash(),
	}
}

func ChoosePubKeyHash() transaction.PubKeyHash {
	keys := data.GetKeys()
	index := rand.Intn(4)
	logger.Debugf("[miner] Chose pubkeyhash at index %d", index)
	return keys[index]
}

func (m *minerService) StartMining(transactions []transaction.Transaction) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cancelMining != nil {
		m.cancelMining()
	}

	if !m.miningEnabled {
		logger.Debugf("[miner] Mining is disabled, ignoring StartMining request")
		return
	}

	tip := m.blockStore.GetMainChainTip()
	previousBlockHash := tip.Hash()
	logger.Infof("[miner] Started mining new block with %d transactions (+1 Coinbase) and PrevBlockHash %v", len(transactions), previousBlockHash)
	candidateBlock, err := m.createCandidateBlock(transactions, m.blockStore.GetCurrentHeight()+1, previousBlockHash)
	if err != nil {
		logger.Errorf("[miner] Failed to create candidate block: %v", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.cancelMining = cancel

	go func() {
		nonce, timestamp, err := m.mineBlock(candidateBlock, ctx)
		if err != nil {
			logger.Infof("[miner] Mining stopped: %v", err)
			return
		}
		candidateBlock.Header.Nonce = nonce
		candidateBlock.Header.Timestamp = timestamp
		logger.Infof("[miner] Mined new block: %v", &candidateBlock.Header)
		m.blockchain.AddSelfMinedBlock(candidateBlock)
	}()
}

func (m *minerService) StopMining() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cancelMining()
}

func (m *minerService) EnableMining() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.miningEnabled {
		logger.Infof("[miner] Mining already enabled")
		return
	}

	logger.Infof("[miner] Enabling mining")
	m.miningEnabled = true
}

func (m *minerService) DisableMining() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.miningEnabled {
		logger.Infof("[miner] Mining already disabled")
		return
	}

	logger.Infof("[miner] Disabling mining")
	m.miningEnabled = false

	// Stop any ongoing mining
	if m.cancelMining != nil {
		logger.Infof("[miner] Stopping ongoing mining due to disable")
		m.cancelMining()
	}
}

// MineBlock Mines a block by change the nonce until the block matches the given difficulty target
func (m *minerService) mineBlock(candidateBlock block.Block, ctx context.Context) (nonce uint32, timestamp int64, err error) {
	target := getTarget(candidateBlock.Header.DifficultyTarget)
	timestamp = candidateBlock.Header.Timestamp

	var counter uint64 = 0
	var hashInt big.Int
	nonce = rand.Uint32()

	for {
		select {
		case <-ctx.Done():
			logger.Infof("[miner] Mining cancelled")
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
