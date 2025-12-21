package infrastructure

import (
	"encoding/binary"
	"errors"
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

	// Key prefixes for different types of entries
	utxoPrefixByte = 0x00 // utxo:<outpoint> -> UTXOEntry
	pkhPrefixByte  = 0x01 // pkh:<pubkeyhash>:<outpoint> -> empty (secondary index)
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

// utxoKey returns the primary key for a UTXO entry: utxo:<outpoint>
func utxoKey(outpoint utxopool.Outpoint) []byte {
	opKey := outpoint.Key()
	key := make([]byte, 1+len(opKey))
	key[0] = utxoPrefixByte
	copy(key[1:], opKey)
	return key
}

// pkhIndexKey returns the secondary index key: pkh:<pubkeyhash>:<outpoint>
func pkhIndexKey(pubKeyHash transaction.PubKeyHash, outpoint utxopool.Outpoint) []byte {
	opKey := outpoint.Key()
	// 1 byte prefix + 20 bytes pubkeyhash + outpoint key
	key := make([]byte, 1+common.PublicKeyHashSize+len(opKey))
	key[0] = pkhPrefixByte
	copy(key[1:1+common.PublicKeyHashSize], pubKeyHash[:])
	copy(key[1+common.PublicKeyHashSize:], opKey)
	return key
}

// pkhIndexPrefix returns the prefix for scanning all outpoints belonging to a PubKeyHash
func pkhIndexPrefix(pubKeyHash transaction.PubKeyHash) []byte {
	prefix := make([]byte, 1+common.PublicKeyHashSize)
	prefix[0] = pkhPrefixByte
	copy(prefix[1:], pubKeyHash[:])
	return prefix
}

// outpointFromPkhIndexKey extracts the outpoint from a pkh index key
func outpointFromPkhIndexKey(key []byte) utxopool.Outpoint {
	// Skip prefix byte + pubkeyhash
	opKeyStart := 1 + common.PublicKeyHashSize
	return utxopool.OutpointFromKey(key[opKeyStart:])
}

func (c *UTXOEntryDAOImpl) Update(outpoint utxopool.Outpoint, entry utxopool.UTXOEntry) error {
	return c.db.Update(func(txn *badger.Txn) error {
		// Set primary UTXO entry
		if err := txn.Set(utxoKey(outpoint), encodeUTXOEntry(entry)); err != nil {
			return err
		}
		// Set secondary PubKeyHash index (empty value)
		return txn.Set(pkhIndexKey(entry.Output.PubKeyHash, outpoint), nil)
	})
}

func (c *UTXOEntryDAOImpl) Delete(outpoint utxopool.Outpoint) error {
	return c.db.Update(func(txn *badger.Txn) error {
		// First, find the entry to get its PubKeyHash for index deletion
		item, err := txn.Get(utxoKey(outpoint))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return nil // Already deleted
			}
			return err
		}

		var entry utxopool.UTXOEntry
		err = item.Value(func(val []byte) error {
			entry = decodeUTXOEntry(val)
			return nil
		})
		if err != nil {
			return err
		}

		// Delete secondary index first
		if err = txn.Delete(pkhIndexKey(entry.Output.PubKeyHash, outpoint)); err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return err
		}

		// Delete primary entry
		return txn.Delete(utxoKey(outpoint))
	})
}

func (c *UTXOEntryDAOImpl) Find(outpoint utxopool.Outpoint) (utxopool.UTXOEntry, error) {
	var entry utxopool.UTXOEntry
	err := c.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(utxoKey(outpoint))
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

// FindByPubKeyHash returns all outpoints belonging to the given PubKeyHash
// using a secondary index prefix scan
func (c *UTXOEntryDAOImpl) FindByPubKeyHash(pubKeyHash transaction.PubKeyHash) ([]utxopool.Outpoint, error) {
	var outpoints []utxopool.Outpoint

	err := c.db.View(func(txn *badger.Txn) error {
		prefix := pkhIndexPrefix(pubKeyHash)
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false // We only need keys
		opts.Prefix = prefix

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			outpoint := outpointFromPkhIndexKey(item.Key())
			outpoints = append(outpoints, outpoint)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return outpoints, nil
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
