package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"slices"
	"sort"
	"time"

	"bjoernblessin.de/go-utils/util/logger"
)

type transactionWithFee struct {
	tx  transaction.Transaction
	Fee uint64
}

const TxPerBlock = 100

func (m *minerService) createCandidateBlock(transactions []transaction.Transaction, height uint64, currentTip common.Hash) (block.Block, error) {
	tx, err := m.buildTransactions(transactions, height, currentTip)
	if err != nil {
		return block.Block{}, err
	}

	header, err := m.createCandidateBlockHeader(tx)
	if err != nil {
		return block.Block{}, err
	}

	return block.Block{Header: header, Transactions: tx}, nil
}

func (m *minerService) createCandidateBlockHeader(transactions []transaction.Transaction) (block.BlockHeader, error) {
	tip := m.blockStore.GetMainChainTip()
	previousBlockHash := tip.Hash()

	merkleRoot := block.MerkleRootFromTransactions(transactions)
	logger.Tracef("[miner] Calculated merkle root: %v", merkleRoot)

	targetBits, err := GetCurrentTargetBits()
	if err != nil {
		return block.BlockHeader{}, err
	}

	blockHeader := block.BlockHeader{
		PreviousBlockHash: previousBlockHash,
		MerkleRoot:        merkleRoot,
		Timestamp:         time.Now().Unix(),
		DifficultyTarget:  targetBits,
		Nonce:             0,
	}
	return blockHeader, nil
}

func (m *minerService) buildTransactions(transactions []transaction.Transaction, height uint64, currentTip common.Hash) ([]transaction.Transaction, error) {
	transactionsWithFees, err := m.getTransactionWithFee(transactions, currentTip)
	if err != nil {
		return nil, err
	}

	transactionsSorted := m.sortAndReversedTransactions(transactionsWithFees)

	coinbaseTx, err := m.createCoinbaseTransaction(transactionsWithFees, height)
	if err != nil {
		return nil, err
	}
	transactionCount := int(min(TxPerBlock-1, float64(len(transactionsSorted))))
	txToPutInBlock := append([]transaction.Transaction{coinbaseTx}, transactionsSorted[:transactionCount]...)
	return txToPutInBlock, nil
}

func (m *minerService) sortAndReversedTransactions(transactionsWithFees []transactionWithFee) []transaction.Transaction {
	sort.Slice(transactionsWithFees[:], func(i, j int) bool {
		return transactionsWithFees[i].Fee < transactionsWithFees[j].Fee
	})
	transactionsSorted := make([]transaction.Transaction, len(transactionsWithFees))
	for i, tx := range transactionsWithFees {
		transactionsSorted[i] = tx.tx
	}
	slices.Reverse(transactionsSorted)
	return transactionsSorted
}

func (m *minerService) createCoinbaseTransaction(transactions []transactionWithFee, height uint64) (transaction.Transaction, error) {
	var sumOfFees uint64
	for _, tx := range transactions {
		sumOfFees += tx.Fee
	}

	reward, err := GetCurrentReward()
	if err != nil {
		return transaction.Transaction{}, err
	}

	coinbaseTransaction := transaction.NewCoinbaseTransaction(m.ownPubKeyHash, sumOfFees+reward, height)

	return coinbaseTransaction, nil
}

func (m *minerService) getTransactionWithFee(transactions []transaction.Transaction, currentTip common.Hash) ([]transactionWithFee, error) {
	transactionsWithFees := make([]transactionWithFee, len(transactions))
	for i, tx := range transactions {
		var inputSum uint64
		inputSum, err := m.getInputSum(tx, currentTip)
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

func (m *minerService) getInputSum(tx transaction.Transaction, currentTip common.Hash) (inputSum uint64, err error) {
	for _, input := range tx.Inputs {
		utxoResult, err := m.utxoService.GetUtxoFromBlock(input.PrevTxID, input.OutputIndex, currentTip)
		if err != nil {
			return 0, err
		}
		inputSum += utxoResult.Value
	}
	return inputSum, nil
}

func GetCurrentTargetBits() (uint8, error) {
	return 28, nil
}

func GetCurrentReward() (uint64, error) {
	return 50, nil
}
