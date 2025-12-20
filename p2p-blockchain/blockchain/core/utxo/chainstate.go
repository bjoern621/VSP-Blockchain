package utxo

import (
	"errors"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/transaction"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/utxopool"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/infrastructure"
	"sync"

	"github.com/dgraph-io/badger/v4"
)

const (
	DefaultCacheSize = 10000
)

// ChainStateService provides persistent UTXO storage with an in-memory cache
// It uses BadgerDB for on-disk storage and an LRU cache for frequently accessed UTXOs
type ChainStateService struct {
	mu  sync.RWMutex
	dao *infrastructure.UTXOEntryDAO
	// cache is an LRU cache for recently accessed UTXOs
	cache     map[string]utxopool.UTXOEntry
	cacheKeys []string // keys in order of access
	cacheSize int
}

// ChainStateConfig holds configuration for the ChainStateService
type ChainStateConfig struct {
	// DBPath is the path to the BadgerDB database directory
	DBPath string

	CacheSize int

	// InMemory used for testing
	InMemory bool
}

// NewChainState creates a new ChainStateService with BadgerDB storage
func NewChainState(config ChainStateConfig) (*ChainStateService, error) {
	daoConfig := infrastructure.NewUTXOEntryDAOConfig(config.DBPath, config.InMemory)
	dao, err := infrastructure.GetUTXOEntryDAO(daoConfig)
	if err != nil {
		return nil, err
	}
	cacheSize := config.CacheSize
	if cacheSize <= 0 {
		cacheSize = DefaultCacheSize
	}

	return &ChainStateService{
		dao:       dao,
		cache:     make(map[string]utxopool.UTXOEntry),
		cacheKeys: make([]string, 0, cacheSize),
		cacheSize: cacheSize,
	}, nil
}

// Get retrieves a UTXO from the chainstate
func (c *ChainStateService) Get(outpoint utxopool.Outpoint) (utxopool.UTXOEntry, error) {
	key := string(outpoint.Key())

	utxoEntry, ok := c.getFromCache(key)
	if ok {
		return utxoEntry, nil
	}

	entry, err := c.dao.Find(outpoint)

	if err != nil {
		return utxopool.UTXOEntry{}, err
	}

	c.addToCache(key, entry)

	return entry, nil
}

func (c *ChainStateService) getFromCache(key string) (utxopool.UTXOEntry, bool) {
	c.mu.RLock()
	if entry, ok := c.cache[key]; ok {
		c.mu.RUnlock()
		c.touchCache(key)
		return entry, true
	}
	c.mu.RUnlock()
	return utxopool.UTXOEntry{}, false
}

// Add adds a new UTXO to the chainstate
func (c *ChainStateService) Add(outpoint utxopool.Outpoint, entry utxopool.UTXOEntry) error {
	err := c.dao.Update(outpoint, entry)
	if err != nil {
		return err
	}

	key := outpoint.Key()
	c.addToCache(string(key), entry)
	return nil
}

// Remove removes a UTXO from the chainstate
func (c *ChainStateService) Remove(outpoint utxopool.Outpoint) error {
	err := c.dao.Delete(outpoint)
	if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
		return err
	}

	key := outpoint.Key()
	c.removeFromCache(string(key))
	return nil
}

// Contains checks if a UTXO exists in the chainstate
func (c *ChainStateService) Contains(outpoint utxopool.Outpoint) bool {
	key := string(outpoint.Key())

	fnd := c.containedInCache(key)
	if fnd {
		return fnd
	}

	return c.containedOnDisk(outpoint)
}

func (c *ChainStateService) containedOnDisk(outpoint utxopool.Outpoint) bool {
	_, err := c.dao.Find(outpoint)
	return err == nil
}

func (c *ChainStateService) containedInCache(key string) bool {
	c.mu.RLock()
	if _, ok := c.cache[key]; ok {
		c.mu.RUnlock()
		return true
	}
	c.mu.RUnlock()
	return false
}

// GetUTXO retrieves an output by transaction ID and output index
func (c *ChainStateService) GetUTXO(txID transaction.TransactionID, outputIndex uint32) (transaction.Output, error) {
	outpoint := utxopool.NewOutpoint(txID, outputIndex)
	entry, err := c.Get(outpoint)
	if err != nil {
		return transaction.Output{}, err
	}
	return entry.Output, err
}

// GetUTXOEntry retrieves the full UTXO entry with metadata
func (c *ChainStateService) GetUTXOEntry(outpoint utxopool.Outpoint) (utxopool.UTXOEntry, error) {
	return c.Get(outpoint)
}

// ContainsUTXO checks if a UTXO exists
func (c *ChainStateService) ContainsUTXO(outpoint utxopool.Outpoint) bool {
	return c.Contains(outpoint)
}

// Close closes the database
func (c *ChainStateService) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = nil
	c.cacheKeys = nil

	return c.dao.Close()
}

// Flush persists pending writes to disk
func (c *ChainStateService) Flush() error {
	return c.dao.Persist()
}

// touchCache moves a key to the end of the LRU list (most recently used)
func (c *ChainStateService) touchCache(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Find and remove the key from its current position
	for i, k := range c.cacheKeys {
		if k == key {
			c.cacheKeys = append(c.cacheKeys[:i], c.cacheKeys[i+1:]...)
			break
		}
	}
	// Add to the end (most recently used)
	c.cacheKeys = append(c.cacheKeys, key)
}

// addToCache adds an entry to the cache, evicting old entries if necessary
func (c *ChainStateService) addToCache(key string, entry utxopool.UTXOEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If key already exists, just update
	if _, exists := c.cache[key]; exists {
		c.cache[key] = entry
		return
	}

	// Evict oldest entries if cache is full
	for len(c.cache) >= c.cacheSize && len(c.cacheKeys) > 0 {
		oldestKey := c.cacheKeys[0]
		c.cacheKeys = c.cacheKeys[1:]
		delete(c.cache, oldestKey)
	}

	c.cache[key] = entry
	c.cacheKeys = append(c.cacheKeys, key)
}

// removeFromCache removes an entry from the cache
func (c *ChainStateService) removeFromCache(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.cache, key)
	for i, k := range c.cacheKeys {
		if k == key {
			c.cacheKeys = append(c.cacheKeys[:i], c.cacheKeys[i+1:]...)
			break
		}
	}
}

// CacheStats returns the current cache statistics
func (c *ChainStateService) CacheStats() (size int, capacity int) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.cache), c.cacheSize
}
