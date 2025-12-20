package utxo

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/transaction"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/utxopool"
)

// LookupService provides a read-only view of UTXOs
type LookupService interface {
	// GetUTXO retrieves an output by transaction ID and output index
	GetUTXO(txID transaction.TransactionID, outputIndex uint32) (transaction.Output, error)

	// GetUTXOEntry retrieves the full UTXO entry with metadata
	GetUTXOEntry(outpoint utxopool.Outpoint) (utxopool.UTXOEntry, error)

	// ContainsUTXO checks if a UTXO exists
	ContainsUTXO(outpoint utxopool.Outpoint) bool
}
