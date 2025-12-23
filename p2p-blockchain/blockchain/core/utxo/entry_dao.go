package utxo

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/utxopool"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
)

// EntryDAO provides data access operations for UTXO entries
type EntryDAO interface {
	// Update adds or updates a UTXO entry at the given outpoint
	Update(outpoint utxopool.Outpoint, entry utxopool.UTXOEntry) error

	// Delete removes a UTXO entry at the given outpoint
	Delete(outpoint utxopool.Outpoint) error

	// Find retrieves a UTXO entry by its outpoint
	Find(outpoint utxopool.Outpoint) (utxopool.UTXOEntry, error)

	// FindByPubKeyHash returns all outpoints belonging to the given PubKeyHash.
	FindByPubKeyHash(pubKeyHash transaction.PubKeyHash) ([]utxopool.Outpoint, error)

	// Close closes the DAO and releases resources
	Close() error

	// Persist ensures all pending writes are persisted to disk
	Persist() error
}
