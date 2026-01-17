package utxo

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/utxopool"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"sync"
)

// UTXOView provides an ephemeral, read-write overlay on top of a base UTXO state.
// It is used for block validation without mutating the global UTXO state.
// Views are short-lived and discarded after validation.
//
// Invariant: Views are ephemeral - never persist state to disk.
// Invariant: Multiple transactions can be applied sequentially within a view.
type UTXOView interface {
	// Get retrieves a UTXO entry, checking overlay first then base.
	Get(outpoint utxopool.Outpoint) (utxopool.UTXOEntry, error)

	// IsSpent checks if an outpoint has been spent in this view.
	IsSpent(outpoint utxopool.Outpoint) bool

	// ApplyTx applies a transaction to this view (spends inputs, creates outputs).
	// This is used for sequential validation of transactions within a block.
	ApplyTx(tx *transaction.Transaction, txID transaction.TransactionID, blockHeight uint64, isCoinbase bool) error

	// GetAddedUTXOs returns all UTXOs added in this view (for delta extraction).
	GetAddedUTXOs() map[string]utxopool.UTXOEntry

	// GetSpentOutpoints returns all outpoints spent in this view (for delta extraction).
	GetSpentOutpoints() map[string]struct{}
}

// BaseUTXOProvider is the interface for the underlying UTXO state.
// This can be the main chain UTXO set or a reconstructed state at a fork point.
type BaseUTXOProvider interface {
	Get(outpoint utxopool.Outpoint) (utxopool.UTXOEntry, error)
	Contains(outpoint utxopool.Outpoint) bool
}

// ephemeralUTXOView is a concrete implementation of UTXOView.
// It overlays changes on top of a base provider without mutating it.
type ephemeralUTXOView struct {
	mu sync.RWMutex

	// base is the underlying UTXO state (main chain or fork point state)
	base BaseUTXOProvider

	// added contains UTXOs created in this view
	added map[string]utxopool.UTXOEntry

	// spent contains outpoints that have been spent in this view
	spent map[string]struct{}
}

// NewEphemeralUTXOView creates a new ephemeral UTXO view on top of a base provider.
func NewEphemeralUTXOView(base BaseUTXOProvider) UTXOView {
	return &ephemeralUTXOView{
		base:  base,
		added: make(map[string]utxopool.UTXOEntry),
		spent: make(map[string]struct{}),
	}
}

// Get retrieves a UTXO entry, checking:
// 1. If spent in this view -> error
// 2. If added in this view -> return it
// 3. Check base provider -> return if found
func (v *ephemeralUTXOView) Get(outpoint utxopool.Outpoint) (utxopool.UTXOEntry, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	key := string(outpoint.Key())

	// Check if spent in this view
	if _, isSpent := v.spent[key]; isSpent {
		return utxopool.UTXOEntry{}, ErrUTXOAlreadySpent
	}

	// Check if added in this view
	if entry, ok := v.added[key]; ok {
		return entry, nil
	}

	// Check base provider
	return v.base.Get(outpoint)
}

// IsSpent checks if an outpoint has been spent in this view.
func (v *ephemeralUTXOView) IsSpent(outpoint utxopool.Outpoint) bool {
	v.mu.RLock()
	defer v.mu.RUnlock()

	key := string(outpoint.Key())
	_, isSpent := v.spent[key]
	return isSpent
}

// ApplyTx applies a transaction to this view:
// 1. Mark all inputs as spent
// 2. Add all outputs to the view
func (v *ephemeralUTXOView) ApplyTx(tx *transaction.Transaction, txID transaction.TransactionID, blockHeight uint64, isCoinbase bool) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	// First, validate and mark inputs as spent
	for _, input := range tx.Inputs {
		outpoint := utxopool.NewOutpoint(input.PrevTxID, input.OutputIndex)
		key := string(outpoint.Key())

		// Check if already spent in this view
		if _, isSpent := v.spent[key]; isSpent {
			return ErrUTXOAlreadySpent
		}

		// Check if it exists (either in added or base)
		_, existsInAdded := v.added[key]
		if !existsInAdded {
			if !v.base.Contains(outpoint) {
				return ErrUTXONotFound
			}
			// Mark as spent only if it was from base (not created in this view)
			v.spent[key] = struct{}{}
		} else {
			// Remove from added if it was added in this view
			// (net effect: created and consumed within same view = nothing)
			delete(v.added, key)
		}
	}

	// Add outputs
	for i, output := range tx.Outputs {
		outpoint := utxopool.NewOutpoint(txID, uint32(i))
		key := string(outpoint.Key())

		entry := utxopool.NewUTXOEntry(output, blockHeight, isCoinbase && i == 0)
		v.added[key] = entry
	}

	return nil
}

// GetAddedUTXOs returns a copy of all UTXOs added in this view.
func (v *ephemeralUTXOView) GetAddedUTXOs() map[string]utxopool.UTXOEntry {
	v.mu.RLock()
	defer v.mu.RUnlock()

	result := make(map[string]utxopool.UTXOEntry, len(v.added))
	for k, entry := range v.added {
		result[k] = entry
	}
	return result
}

// GetSpentOutpoints returns a copy of all outpoints spent in this view.
func (v *ephemeralUTXOView) GetSpentOutpoints() map[string]struct{} {
	v.mu.RLock()
	defer v.mu.RUnlock()

	result := make(map[string]struct{}, len(v.spent))
	for k := range v.spent {
		result[k] = struct{}{}
	}
	return result
}

// chainStateBaseProvider wraps ChainStateService to implement BaseUTXOProvider.
type chainStateBaseProvider struct {
	chainstate *ChainStateService
}

// NewChainStateBaseProvider creates a BaseUTXOProvider from a ChainStateService.
func NewChainStateBaseProvider(chainstate *ChainStateService) BaseUTXOProvider {
	return &chainStateBaseProvider{chainstate: chainstate}
}

func (p *chainStateBaseProvider) Get(outpoint utxopool.Outpoint) (utxopool.UTXOEntry, error) {
	return p.chainstate.Get(outpoint)
}

func (p *chainStateBaseProvider) Contains(outpoint utxopool.Outpoint) bool {
	return p.chainstate.Contains(outpoint)
}

// deltaBaseProvider wraps a base provider and applies a delta on top.
// This allows creating views on side chain state without a full UTXO set.
type deltaBaseProvider struct {
	base  BaseUTXOProvider
	delta *SideChainDelta
}

// NewDeltaBaseProvider creates a BaseUTXOProvider that applies a delta on top of a base.
func NewDeltaBaseProvider(base BaseUTXOProvider, delta *SideChainDelta) BaseUTXOProvider {
	return &deltaBaseProvider{
		base:  base,
		delta: delta,
	}
}

func (p *deltaBaseProvider) Get(outpoint utxopool.Outpoint) (utxopool.UTXOEntry, error) {
	key := string(outpoint.Key())

	// Check if spent in delta
	if _, isSpent := p.delta.SpentUTXOs[key]; isSpent {
		return utxopool.UTXOEntry{}, ErrUTXOAlreadySpent
	}

	// Check if added in delta
	if entry, exists := p.delta.AddedUTXOs[key]; exists {
		return entry, nil
	}

	// Fall back to base
	return p.base.Get(outpoint)
}

func (p *deltaBaseProvider) Contains(outpoint utxopool.Outpoint) bool {
	key := string(outpoint.Key())

	// Check if spent in delta
	if _, isSpent := p.delta.SpentUTXOs[key]; isSpent {
		return false
	}

	// Check if added in delta
	if _, exists := p.delta.AddedUTXOs[key]; exists {
		return true
	}

	// Fall back to base
	return p.base.Contains(outpoint)
}

// chainedDeltaBaseProvider allows stacking multiple deltas on top of each other.
// This is useful for branches off of side chains.
type chainedDeltaBaseProvider struct {
	base   BaseUTXOProvider
	deltas []*SideChainDelta // deltas[0] is closest to base, deltas[len-1] is most recent
}

// NewChainedDeltaBaseProvider creates a BaseUTXOProvider that applies multiple deltas.
// Deltas should be ordered from oldest to newest (fork point to tip).
func NewChainedDeltaBaseProvider(base BaseUTXOProvider, deltas []*SideChainDelta) BaseUTXOProvider {
	if len(deltas) == 0 {
		return base
	}
	if len(deltas) == 1 {
		return NewDeltaBaseProvider(base, deltas[0])
	}
	return &chainedDeltaBaseProvider{
		base:   base,
		deltas: deltas,
	}
}

func (p *chainedDeltaBaseProvider) Get(outpoint utxopool.Outpoint) (utxopool.UTXOEntry, error) {
	key := string(outpoint.Key())

	// Check deltas from newest to oldest
	for i := len(p.deltas) - 1; i >= 0; i-- {
		delta := p.deltas[i]

		// Check if spent in this delta
		if _, isSpent := delta.SpentUTXOs[key]; isSpent {
			return utxopool.UTXOEntry{}, ErrUTXOAlreadySpent
		}

		// Check if added in this delta
		if entry, exists := delta.AddedUTXOs[key]; exists {
			return entry, nil
		}
	}

	// Fall back to base
	return p.base.Get(outpoint)
}

func (p *chainedDeltaBaseProvider) Contains(outpoint utxopool.Outpoint) bool {
	key := string(outpoint.Key())

	// Check deltas from newest to oldest
	for i := len(p.deltas) - 1; i >= 0; i-- {
		delta := p.deltas[i]

		// Check if spent in this delta
		if _, isSpent := delta.SpentUTXOs[key]; isSpent {
			return false
		}

		// Check if added in this delta
		if _, exists := delta.AddedUTXOs[key]; exists {
			return true
		}
	}

	// Fall back to base
	return p.base.Contains(outpoint)
}
