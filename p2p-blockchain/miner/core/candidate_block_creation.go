package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"slices"
	"sort"
	"time"
)

type transactionWithFee struct {
	tx  transaction.Transaction
	Fee uint64
}

const TxPerBlock = 100

func (m *minerService) createCandidateBlock(transactions []transaction.Transaction) (block.Block, error) {
	tx, err := m.buildTransactions(transactions)
	if err != nil {
		return block.Block{}, err
	}

	header, err := m.createCandidateBlockHeader(transactions)
	if err != nil {
		return block.Block{}, err
	}

	return block.Block{Header: header, Transactions: tx}, nil
}

func (m *minerService) createCandidateBlockHeader(transactions []transaction.Transaction) (block.BlockHeader, error) {
	tip := m.blockStore.GetMainChainTip()
	previousBlockHash := tip.Hash()

	merkleRoot := block.MerkleRootFromTransactions(transactions)

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

func (m *minerService) buildTransactions(transactions []transaction.Transaction) ([]transaction.Transaction, error) {
	transactionsWithFees, err := m.getTransactionWithFee(transactions)
	if err != nil {
		return nil, err
	}

	transactionsSorted := m.sortAndReversedTransactions(transactionsWithFees)

	coinbaseTx, err := m.createCoinbaseTransaction(transactionsWithFees)
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

func (m *minerService) createCoinbaseTransaction(transactions []transactionWithFee) (transaction.Transaction, error) {
	var sumOfFees uint64
	for _, tx := range transactions {
		sumOfFees += tx.Fee
	}

	reward, err := GetCurrentReward()
	if err != nil {
		return transaction.Transaction{}, err
	}
	ownPubKeyHash, err := getOwnPubKeyHash()
	if err != nil {
		return transaction.Transaction{}, err
	}

	coinbaseTransaction := transaction.NewCoinbaseTransaction(ownPubKeyHash, sumOfFees+reward, []byte("Hello World"))

	return coinbaseTransaction, nil
}

func (m *minerService) getTransactionWithFee(transactions []transaction.Transaction) ([]transactionWithFee, error) {
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

func (m *minerService) getInputSum(tx transaction.Transaction) (inputSum uint64, err error) {
	for _, input := range tx.Inputs {
		utxoResult, err := m.utxoService.GetUTXO(input.PrevTxID, input.OutputIndex)
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

func getOwnPubKeyHash() (transaction.PubKeyHash, error) {
	return transaction.PubKeyHash{}, nil //TODO
}
