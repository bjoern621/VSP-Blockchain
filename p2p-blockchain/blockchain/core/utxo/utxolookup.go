package utxo

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/utxopool"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
)

// LookupService provides a read-only view of UTXOs
type LookupService interface {
	// GetUTXO retrieves an output by transaction ID and output index
	GetUTXO(txID transaction.TransactionID, outputIndex uint32) (transaction.Output, error)

	// GetUTXOEntry retrieves the full UTXO entry with metadata
	GetUTXOEntry(outpoint utxopool.Outpoint) (utxopool.UTXOEntry, error)

	// ContainsUTXO checks if a UTXO exists
	ContainsUTXO(outpoint utxopool.Outpoint) bool

	// GetUTXOsByPubKeyHash returns all UTXOs belonging to the given PubKeyHash.
	// Results include both confirmed (chainstate) and unconfirmed (mempool) UTXOs,
	// excluding any that have been marked as spent in the mempool.
	// Each result includes the outpoint for proper identification.
	GetUTXOsByPubKeyHash(pubKeyHash transaction.PubKeyHash) ([]transaction.UTXO, error)
}
