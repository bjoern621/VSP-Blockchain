package utxopool

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
)

// ConfirmationDepth is the number of blocks required for a transaction to be considered confirmed.
// A UTXO is confirmed when currentHeight - blockHeight >= ConfirmationDepth.
const ConfirmationDepth = 5

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

// IsConfirmedAt returns true if this UTXO is confirmed at the given chain height.
// A UTXO is confirmed when currentHeight - blockHeight >= ConfirmationDepth.
func (e UTXOEntry) IsConfirmedAt(currentHeight uint64) bool {
	if e.BlockHeight == 0 {
		return false // Unconfirmed mempool transaction
	}
	return currentHeight >= e.BlockHeight+ConfirmationDepth
}

// UTXOWithOutpoint pairs a UTXO entry with its outpoint for lookup results
type UTXOWithOutpoint struct {
	Outpoint Outpoint
	Entry    UTXOEntry
}
