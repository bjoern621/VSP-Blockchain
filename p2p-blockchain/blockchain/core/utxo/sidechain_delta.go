package utxo

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/utxopool"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"sync"
)

// SideChainDelta stores the UTXO differences from a fork point for a side chain.
// Instead of maintaining a full UTXO set per side chain, we only store the delta.
// This is memory-efficient and allows fast side chain validation.
//
// Invariant: SideChainDelta is immutable after creation during block validation.
// Invariant: Deltas are additive - a chain of deltas can be composed.
type SideChainDelta struct {
	// ForkPoint is the block hash where this side chain diverges from the main chain.
	ForkPoint common.Hash

	// ChainTip is the current tip of this side chain.
	ChainTip common.Hash

	// ParentTip is the previous tip of this side chain (before the last block was applied).
	// For the first block after fork point, this equals ForkPoint.
	// This enables delta chaining and fork reconstruction.
	ParentTip common.Hash

	// AddedUTXOs contains UTXOs created in this side chain since the fork point.
	// Key is the outpoint key (string).
	AddedUTXOs map[string]utxopool.UTXOEntry

	// SpentUTXOs contains outpoints spent in this side chain since the fork point.
	// Key is the outpoint key (string).
	SpentUTXOs map[string]struct{}

	// BlockCount is the number of blocks in this side chain from fork point to tip.
	BlockCount uint64
}

// NewSideChainDelta creates a new empty side chain delta.
func NewSideChainDelta(forkPoint, chainTip common.Hash) *SideChainDelta {
	return &SideChainDelta{
		ForkPoint:  forkPoint,
		ChainTip:   chainTip,
		ParentTip:  forkPoint, // Initially, parent is the fork point
		AddedUTXOs: make(map[string]utxopool.UTXOEntry),
		SpentUTXOs: make(map[string]struct{}),
		BlockCount: 0,
	}
}

// NewSideChainDeltaWithParent creates a new empty side chain delta with explicit parent.
func NewSideChainDeltaWithParent(forkPoint, chainTip, parentTip common.Hash) *SideChainDelta {
	return &SideChainDelta{
		ForkPoint:  forkPoint,
		ChainTip:   chainTip,
		ParentTip:  parentTip,
		AddedUTXOs: make(map[string]utxopool.UTXOEntry),
		SpentUTXOs: make(map[string]struct{}),
		BlockCount: 0,
	}
}

// ApplyView merges changes from a validated UTXOView into this delta.
// This is called after a side chain block is successfully validated.
func (d *SideChainDelta) ApplyView(view UTXOView, newTip common.Hash) {
	// Update parent tip to current tip before moving to new tip
	d.ParentTip = d.ChainTip

	// Merge spent outpoints
	for key := range view.GetSpentOutpoints() {
		// If this was added in the delta, remove it (net effect: nothing)
		if _, existsInAdded := d.AddedUTXOs[key]; existsInAdded {
			delete(d.AddedUTXOs, key)
		} else {
			// Mark as spent from main chain state
			d.SpentUTXOs[key] = struct{}{}
		}
	}

	// Merge added UTXOs
	for key, entry := range view.GetAddedUTXOs() {
		d.AddedUTXOs[key] = entry
	}

	d.ChainTip = newTip
	d.BlockCount++
}

// Clone creates a deep copy of the delta for extending side chains.
func (d *SideChainDelta) Clone() *SideChainDelta {
	clone := &SideChainDelta{
		ForkPoint:  d.ForkPoint,
		ChainTip:   d.ChainTip,
		ParentTip:  d.ParentTip,
		AddedUTXOs: make(map[string]utxopool.UTXOEntry, len(d.AddedUTXOs)),
		SpentUTXOs: make(map[string]struct{}, len(d.SpentUTXOs)),
		BlockCount: d.BlockCount,
	}

	for k, v := range d.AddedUTXOs {
		clone.AddedUTXOs[k] = v
	}
	for k := range d.SpentUTXOs {
		clone.SpentUTXOs[k] = struct{}{}
	}

	return clone
}

// sideChainDeltaStore manages all side chain deltas.
// It is thread-safe for concurrent access.
type sideChainDeltaStore struct {
	mu sync.RWMutex

	// deltas maps chain tip hash to its delta
	deltas map[common.Hash]*SideChainDelta

	// forkPointIndex maps fork point hash to all side chains forking from it
	// This is used for efficient pruning and lookup
	forkPointIndex map[common.Hash][]common.Hash

	// parentIndex maps parent tip to child tips (for finding branches)
	parentIndex map[common.Hash][]common.Hash
}

// NewSideChainDeltaStore creates a new side chain delta store.
func NewSideChainDeltaStore() *sideChainDeltaStore {
	return &sideChainDeltaStore{
		deltas:         make(map[common.Hash]*SideChainDelta),
		forkPointIndex: make(map[common.Hash][]common.Hash),
		parentIndex:    make(map[common.Hash][]common.Hash),
	}
}

// Get retrieves a delta by chain tip hash.
func (s *sideChainDeltaStore) Get(chainTip common.Hash) (*SideChainDelta, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	delta, ok := s.deltas[chainTip]
	return delta, ok
}

// GetByParent retrieves all deltas that extend from the given parent tip.
func (s *sideChainDeltaStore) GetByParent(parentTip common.Hash) []*SideChainDelta {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tips := s.parentIndex[parentTip]
	result := make([]*SideChainDelta, 0, len(tips))
	for _, tip := range tips {
		if delta, exists := s.deltas[tip]; exists {
			result = append(result, delta)
		}
	}
	return result
}

// Put stores a delta, replacing any existing one for the same chain tip.
func (s *sideChainDeltaStore) Put(delta *SideChainDelta) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove old index entries if updating existing delta
	if old, exists := s.deltas[delta.ChainTip]; exists {
		s.removeFromForkIndex(old.ForkPoint, delta.ChainTip)
		s.removeFromParentIndex(old.ParentTip, delta.ChainTip)
	}

	s.deltas[delta.ChainTip] = delta

	// Update fork point index
	s.forkPointIndex[delta.ForkPoint] = append(s.forkPointIndex[delta.ForkPoint], delta.ChainTip)

	// Update parent index
	s.parentIndex[delta.ParentTip] = append(s.parentIndex[delta.ParentTip], delta.ChainTip)
}

// Remove removes a delta by chain tip hash.
func (s *sideChainDeltaStore) Remove(chainTip common.Hash) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if delta, exists := s.deltas[chainTip]; exists {
		s.removeFromForkIndex(delta.ForkPoint, chainTip)
		s.removeFromParentIndex(delta.ParentTip, chainTip)
		delete(s.deltas, chainTip)
	}
}

// GetByForkPoint returns all side chain tips forking from the given point.
func (s *sideChainDeltaStore) GetByForkPoint(forkPoint common.Hash) []common.Hash {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tips := s.forkPointIndex[forkPoint]
	result := make([]common.Hash, len(tips))
	copy(result, tips)
	return result
}

// GetAll returns all stored deltas.
func (s *sideChainDeltaStore) GetAll() []*SideChainDelta {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*SideChainDelta, 0, len(s.deltas))
	for _, delta := range s.deltas {
		result = append(result, delta)
	}
	return result
}

// PruneBefore removes all deltas whose fork point is in the given set.
// This is used for safe pruning of old side chains.
func (s *sideChainDeltaStore) PruneBefore(ancientForkPoints map[common.Hash]struct{}) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	pruned := 0
	for forkPoint := range ancientForkPoints {
		tips := s.forkPointIndex[forkPoint]
		for _, tip := range tips {
			if delta, exists := s.deltas[tip]; exists {
				s.removeFromParentIndex(delta.ParentTip, tip)
			}
			delete(s.deltas, tip)
			pruned++
		}
		delete(s.forkPointIndex, forkPoint)
	}
	return pruned
}

// Size returns the number of side chain deltas stored.
func (s *sideChainDeltaStore) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.deltas)
}

// removeFromForkIndex removes a chain tip from the fork point index.
func (s *sideChainDeltaStore) removeFromForkIndex(forkPoint, chainTip common.Hash) {
	tips := s.forkPointIndex[forkPoint]
	for i, tip := range tips {
		if tip == chainTip {
			s.forkPointIndex[forkPoint] = append(tips[:i], tips[i+1:]...)
			break
		}
	}
	if len(s.forkPointIndex[forkPoint]) == 0 {
		delete(s.forkPointIndex, forkPoint)
	}
}

// removeFromParentIndex removes a chain tip from the parent index.
func (s *sideChainDeltaStore) removeFromParentIndex(parentTip, chainTip common.Hash) {
	tips := s.parentIndex[parentTip]
	for i, tip := range tips {
		if tip == chainTip {
			s.parentIndex[parentTip] = append(tips[:i], tips[i+1:]...)
			break
		}
	}
	if len(s.parentIndex[parentTip]) == 0 {
		delete(s.parentIndex, parentTip)
	}
}
