package miner

import (
	"context"
	"fmt"
	"math/big"
	blockchainApi "s3b/vsp-blockchain/p2p-blockchain/blockchain/api"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/utxo"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api"

	"bjoernblessin.de/go-utils/util/logger"
	"slices"
	"sort"
	"time"

	"bjoernblessin.de/go-utils/util/assert"
)

type MinerAPI interface {
	StartMining(transactions []transaction.Transaction) (minedBlock block.Block)
	StopMining()
	createCandidateBlock(transactions []transaction.Transaction) (block.Block, error)
	MineBlock(candidateBlock block.Block) uint32
}

type MinerService struct {
	utxoLookupService utxo.LookupAPI

	cancelMining        context.CancelFunc
	blockchainMsgSender api.BlockchainAPI
	blockchain          blockchainApi.BlockchainAPI
}

func NewMinerService(
	blockchainMsgSender api.BlockchainAPI,
	blockchain blockchainApi.BlockchainAPI,
) *MinerService {
	return &MinerService{
		blockchainMsgSender: blockchainMsgSender,
		blockchain:          blockchain,
	}
}

type transactionWithFee struct {
	tx  transaction.Transaction
	Fee uint64
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

func (m *MinerService) createCandidateBlock(transactions []transaction.Transaction) (block.Block, error) {
	transactionsWithFees, err := m.getTransactionWithFee(transactions)
	if err != nil {
		return block.Block{}, err
	}
	coinbaseTx, err := m.createCoinbaseTransaction(transactionsWithFees)
	if err != nil {
		return block.Block{}, err
	}

	sort.Slice(transactionsWithFees[:], func(i, j int) bool {
		return transactionsWithFees[i].Fee < transactionsWithFees[j].Fee
	})
	transactionsSorted := make([]transaction.Transaction, len(transactionsWithFees))
	for i, tx := range transactionsWithFees {
		transactionsSorted[i] = tx.tx
	}
	slices.Reverse(transactionsSorted)

	txToPutInBlock := append([]transaction.Transaction{coinbaseTx}, transactionsSorted[:MAX_TX_PER_BLOCK-1]...)

	header := createCandidateBlockHeader(txToPutInBlock)

	return block.Block{Header: header}, nil
}

func (m *MinerService) createCoinbaseTransaction(transactions []transactionWithFee) (transaction.Transaction, error) {
	var sumOfFees uint64
	for _, tx := range transactions {
		sumOfFees += tx.Fee
	}
}

func (m *MinerService) StopMining() {
	m.cancelMining()
}

func (m *MinerService) createCandidateBlock(transactions []transaction.Transaction) block.Block {
	header := createCandidateBlockHeader()

	//TODO: Implement
	bounty := getCurrentBounty()
	ownPubKeyHash := getOwnPubKeyHash()
	ownPrivKey := getOwnPrivKey()
	//TODO: Mit bjarne sprechen
	tx, err := transaction.NewTransaction(nil, ownPubKeyHash, sumOfFees+bounty, 0, ownPrivKey)
	if err != nil {
		return transaction.Transaction{}, err
	}
	return *tx, nil
}

func (m *MinerService) getTransactionWithFee(transactions []transaction.Transaction) ([]transactionWithFee, error) {
	transactionsWithFees := make([]transactionWithFee, len(transactions))
	for i, tx := range transactions {
		var inputSum uint64
		inputSum, err := m.getInputSum(tx)
		if err != nil {
			return nil, err
		}
		var outputSum uint64
		for _, output := range tx.Outputs {
			outputSum += output.Value
		}
		transactionsWithFees[i] = transactionWithFee{tx, inputSum - outputSum}
	}

	return transactionsWithFees, nil
}

func (m *MinerService) getInputSum(tx transaction.Transaction) (inputSum uint64, err error) {
	for _, input := range tx.Inputs {
		utxo, err := m.utxoLookupService.GetUTXO(input.PrevTxID, input.OutputIndex)
		if err != nil {
			return 0, err
		}
		inputSum += utxo.Value
	}
	return inputSum, nil
}

// TODO: Kapitel Die Block-Header aufbauen
func createCandidateBlockHeader(transactions []transaction.Transaction) block.BlockHeader {
	target, err := GetCurrentTargetBits()
	assert.IsNotNil(err)

	highestBlockHash := getHighestBlock().Hash() //TODO: Implement this
	merkleRoot := block.MerkleRoot(transactions)

	currentTime := time.Now().Unix()

	return block.BlockHeader{
		PreviousBlockHash: highestBlockHash,
		MerkleRoot:        merkleRoot,
		Timestamp:         currentTime,
		DifficultyTarget:  target,
		Nonce:             0,
	}
}

func GetCurrentTargetBits() (uint32, error) {
	return 5, nil
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
