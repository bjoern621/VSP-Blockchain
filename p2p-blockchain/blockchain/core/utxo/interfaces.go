package utxo

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/transaction"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/utxopool"
)

// Modifier provides write operations on UTXOs
type Modifier interface {
	LookupService

	// AddUTXO adds a new UTXO from a transaction output
	AddUTXO(outpoint utxopool.Outpoint, entry utxopool.UTXOEntry) error

	// SpendUTXO marks a UTXO as spent
	SpendUTXO(outpoint utxopool.Outpoint) error

	// Remove removes a UTXO from the pool
	Remove(outpoint utxopool.Outpoint) error
}

// UTXOService combines both view and modification capabilities and supports transaction-level operations
type UTXOService interface {
	Modifier

	// ApplyTransaction applies a transaction to the UTXO set
	// - Removes all inputs from the pool (marks as spent)
	// - Adds all outputs to the pool
	ApplyTransaction(tx *transaction.Transaction, txID transaction.TransactionID, blockHeight uint64, isCoinbase bool) error

	// RevertTransaction reverts a transaction from the UTXO set
	// - Adds back all inputs to the pool
	// - Removes all outputs from the pool
	RevertTransaction(tx *transaction.Transaction, txID transaction.TransactionID, inputUTXOs []utxopool.UTXOEntry) error

	// Flush persists any pending changes (for stores with write-back caches)
	Flush() error

	// Close closes the pool and releases resources
	Close() error
}
