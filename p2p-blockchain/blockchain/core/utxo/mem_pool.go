package utxo

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/transaction"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/utxopool"
	"sync"
)

// MemPoolService is an in-memory UTXO pool for unconfirmed transactions
// It tracks both created UTXOs (from transaction outputs) and spent UTXOs (from transaction inputs)
type MemPoolService struct {
	mu sync.RWMutex

	// utxos stores unconfirmed UTXOs created by mempool transactions
	utxos map[string]utxopool.UTXOEntry

	// spent tracks UTXOs that have been spent by mempool transactions
	// These outpoints should be considered "not available" even if they exist in chainstate
	spent map[string]struct{}
}

// NewMemUTXOPool creates a new in-memory UTXO pool
func NewMemUTXOPool() *MemPoolService {
	return &MemPoolService{
		utxos: make(map[string]utxopool.UTXOEntry),
		spent: make(map[string]struct{}),
	}
}

// Get retrieves a UTXO from the mempool
func (m *MemPoolService) Get(outpoint utxopool.Outpoint) (utxopool.UTXOEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	entry, ok := m.utxos[string(outpoint.Key())]
	if !ok {
		return utxopool.UTXOEntry{}, ErrUTXONotFound
	}
	return entry, nil
}

// IsSpent checks if an outpoint has been marked as spent in the mempool
func (m *MemPoolService) IsSpent(outpoint utxopool.Outpoint) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, spent := m.spent[string(outpoint.Key())]
	return spent
}

// Add adds a new UTXO to the mempool
func (m *MemPoolService) Add(outpoint utxopool.Outpoint, entry utxopool.UTXOEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := string(outpoint.Key())
	m.utxos[key] = entry
	// If this was previously marked as spent, unmark it
	delete(m.spent, key)
	return nil
}

// Remove removes a UTXO from the mempool
func (m *MemPoolService) Remove(outpoint utxopool.Outpoint) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := string(outpoint.Key())
	delete(m.utxos, key)
	return nil
}

// MarkSpent marks an outpoint as spent in the mempool
// This is used when a mempool transaction spends a chainstate UTXO
func (m *MemPoolService) MarkSpent(outpoint utxopool.Outpoint) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := string(outpoint.Key())
	// First check if this UTXO exists in mempool, if so remove it
	if _, exists := m.utxos[key]; exists {
		delete(m.utxos, key)
	} else {
		// Mark as spent (it's a chainstate UTXO being spent)
		m.spent[key] = struct{}{}
	}
}

// UnmarkSpent removes the spent marker from an outpoint
func (m *MemPoolService) UnmarkSpent(outpoint utxopool.Outpoint) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := string(outpoint.Key())
	delete(m.spent, key)
}

// Contains checks if a UTXO exists in the mempool
func (m *MemPoolService) Contains(outpoint utxopool.Outpoint) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.utxos[string(outpoint.Key())]
	return ok
}

// GetUTXO retrieves an output by transaction ID and output index
func (m *MemPoolService) GetUTXO(txID transaction.TransactionID, outputIndex uint32) (transaction.Output, error) {
	outpoint := utxopool.NewOutpoint(txID, outputIndex)
	entry, err := m.Get(outpoint)
	if err != nil {
		return transaction.Output{}, err
	}
	return entry.Output, nil
}

// GetUTXOEntry retrieves the full UTXO entry with metadata
func (m *MemPoolService) GetUTXOEntry(outpoint utxopool.Outpoint) (utxopool.UTXOEntry, error) {
	return m.Get(outpoint)
}

// ContainsUTXO checks if a UTXO exists
func (m *MemPoolService) ContainsUTXO(outpoint utxopool.Outpoint) bool {
	return m.Contains(outpoint)
}

// Clear removes all UTXOs and spent markers from the mempool
func (m *MemPoolService) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.utxos = make(map[string]utxopool.UTXOEntry)
	m.spent = make(map[string]struct{})
}

// Size returns the number of UTXOs in the mempool
func (m *MemPoolService) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.utxos)
}

// SpentSize returns the number of spent markers in the mempool
func (m *MemPoolService) SpentSize() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.spent)
}

// Close releases resources (no-op for in-memory store)
func (m *MemPoolService) Close() error {
	m.Clear()
	return nil
}
