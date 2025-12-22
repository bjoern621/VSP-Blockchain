package utxopool

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
)

// UTXOEntry represents a stored UTXO with metadata
type UTXOEntry struct {
	Output      transaction.Output
	BlockHeight uint64 // 0 for unconfirmed transactions
	IsCoinbase  bool
}

// NewUTXOEntry creates a new UTXO entry
func NewUTXOEntry(output transaction.Output, blockHeight uint64, isCoinbase bool) UTXOEntry {
	return UTXOEntry{
		Output:      output,
		BlockHeight: blockHeight,
		IsCoinbase:  isCoinbase,
	}
}

// IsConfirmed returns true if this UTXO is from a confirmed block
func (e UTXOEntry) IsConfirmed() bool {
	return e.BlockHeight > 5
}

// UTXOWithOutpoint pairs a UTXO entry with its outpoint for lookup results
type UTXOWithOutpoint struct {
	Outpoint Outpoint
	Entry    UTXOEntry
}
