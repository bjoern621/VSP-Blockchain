// Package block provides data structures and functions for managing blocks in the blockchain.
package block

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/transaction"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api/blockchain/dto"
)

// Block represents a block in the blockchain.
// It consists of a BlockHeader and a list of Transactions.
type Block struct {
	Header BlockHeader
	// Transactions holds all transactions included in the block.
	// The first transaction is always the coinbase transaction.
	// This slice should never be empty.
	Transactions []transaction.Transaction
}

func NewBlockFromDTO(b dto.BlockDTO) Block {
	txs := make([]transaction.Transaction, 0, len(b.Transactions))
	for i := range b.Transactions {
		txs = append(txs, transaction.NewTransactionFromDTO(b.Transactions[i]))
	}
	return Block{
		Header:       NewBlockHeaderFromDTO(b.Header),
		Transactions: txs,
	}
}
