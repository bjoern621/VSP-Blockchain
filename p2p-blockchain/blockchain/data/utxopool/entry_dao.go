package utxopool

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
)

// EntryDAO provides data access operations for UTXO entries
type EntryDAO interface {
	// Update adds or updates a UTXO entry at the given outpoint
	Update(outpoint Outpoint, entry UTXOEntry) error

	// Delete removes a UTXO entry at the given outpoint
	Delete(outpoint Outpoint) error

	// Find retrieves a UTXO entry by its outpoint
	Find(outpoint Outpoint) (UTXOEntry, error)

	// FindByPubKeyHash returns all outpoints belonging to the given PubKeyHash.
	FindByPubKeyHash(pubKeyHash transaction.PubKeyHash) ([]Outpoint, error)

	// Close closes the DAO and releases resources
	Close() error

	// Persist ensures all pending writes are persisted to disk
	Persist() error
}
