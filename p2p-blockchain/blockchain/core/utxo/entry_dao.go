package utxo

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/utxopool"
)

type EntryDAO interface {
	Update(outpoint utxopool.Outpoint, entry utxopool.UTXOEntry) error

	Delete(outpoint utxopool.Outpoint) error

	Find(outpoint utxopool.Outpoint) (utxopool.UTXOEntry, error)

	Close() error

	Persist() error
}
