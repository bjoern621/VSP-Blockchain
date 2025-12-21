package infrastructure

import (
	"encoding/binary"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/transaction"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/utxopool"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"

	"github.com/dgraph-io/badger/v4"
)

const (
	ValueSize        = 8
	BlockHeightSize  = 8
	IsCoinbaseSize   = 1
	entryEncodedSize = ValueSize + common.PublicKeyHashSize + BlockHeightSize + IsCoinbaseSize
)

type UTXOEntryDAOImpl struct {
	db *badger.DB
}

// NewUTXOEntryDAO returns the singleton instance of UTXOEntryDAOImpl
func NewUTXOEntryDAO(config UTXOEntryDAOConfig) (*UTXOEntryDAOImpl, error) {
	opts := badger.DefaultOptions(config.DBPath)
	if config.InMemory {
		opts = opts.WithInMemory(true)
	}
	opts = opts.WithLogger(nil)

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &UTXOEntryDAOImpl{db: db}, nil
}

func (c *UTXOEntryDAOImpl) Update(outpoint utxopool.Outpoint, entry utxopool.UTXOEntry) error {
	key := outpoint.Key()

	return c.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, encodeUTXOEntry(entry))
	})
}

func (c *UTXOEntryDAOImpl) Delete(outpoint utxopool.Outpoint) error {
	key := outpoint.Key()
	return c.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

func (c *UTXOEntryDAOImpl) Find(outpoint utxopool.Outpoint) (utxopool.UTXOEntry, error) {
	var entry utxopool.UTXOEntry
	err := c.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(outpoint.Key())
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			entry = decodeUTXOEntry(val)
			return nil
		})
	})
	return entry, err
}

func (c *UTXOEntryDAOImpl) Close() error {
	return c.db.Close()
}

func (c *UTXOEntryDAOImpl) Persist() error {
	return c.db.Sync()
}

type UTXOEntryDAOConfig struct {
	DBPath   string
	InMemory bool
}

func NewUTXOEntryDAOConfig(dbPath string, inMemory bool) UTXOEntryDAOConfig {
	return UTXOEntryDAOConfig{
		DBPath:   dbPath,
		InMemory: inMemory,
	}
}

// encodeUTXOEntry encodes a UTXOEntry to bytes
func encodeUTXOEntry(entry utxopool.UTXOEntry) []byte {
	data := make([]byte, entryEncodedSize)
	offset := 0

	binary.BigEndian.PutUint64(data[offset:], entry.Output.Value)
	offset += ValueSize

	copy(data[offset:], entry.Output.PubKeyHash[:])
	offset += common.PublicKeyHashSize

	binary.BigEndian.PutUint64(data[offset:], entry.BlockHeight)
	offset += BlockHeightSize

	if entry.IsCoinbase {
		data[offset] = 1
	} else {
		data[offset] = 0
	}

	return data
}

// decodeUTXOEntry decodes a UTXOEntry from bytes
func decodeUTXOEntry(data []byte) utxopool.UTXOEntry {
	offset := 0

	value := binary.BigEndian.Uint64(data[offset:])
	offset += ValueSize

	var pubKeyHash transaction.PubKeyHash
	copy(pubKeyHash[:], data[offset:offset+common.PublicKeyHashSize])
	offset += common.PublicKeyHashSize

	blockHeight := binary.BigEndian.Uint64(data[offset:])
	offset += BlockHeightSize

	isCoinbase := data[offset] == 1

	return utxopool.UTXOEntry{
		Output: transaction.Output{
			Value:      value,
			PubKeyHash: pubKeyHash,
		},
		BlockHeight: blockHeight,
		IsCoinbase:  isCoinbase,
	}
}
