package utxo

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/utxopool"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
)

// MultiChainLookupService provides UTXO lookups for both main chain and side chains.
// It wraps MultiChainUTXOService and implements LookupService for main chain lookups,
// while also providing methods for side chain lookups.
//
// Usage patterns:
// - For main chain queries (wallets, balance checks): use LookupService methods
// - For side chain queries: use LookupAtChainTip methods
// - For block validation: use the underlying MultiChainUTXOService directly
type MultiChainLookupService struct {
	multiChain *MultiChainUTXOService
}

// Verify MultiChainLookupService implements LookupService
var _ LookupService = (*MultiChainLookupService)(nil)

// NewMultiChainLookupService creates a new multi-chain lookup service.
func NewMultiChainLookupService(multiChain *MultiChainUTXOService) *MultiChainLookupService {
	return &MultiChainLookupService{
		multiChain: multiChain,
	}
}

// =========================================================================
// LookupService Implementation (Main Chain)
// =========================================================================

// GetUTXO retrieves an output from the main chain by transaction ID and output index.
// This looks up UTXOs from both chainstate (confirmed) and mempool (unconfirmed).
func (s *MultiChainLookupService) GetUTXO(txID transaction.TransactionID, outputIndex uint32) (transaction.Output, error) {
	return s.multiChain.GetMainChain().GetUTXO(txID, outputIndex)
}

// GetUTXOEntry retrieves the full UTXO entry from the main chain with metadata.
// This looks up UTXOs from both chainstate (confirmed) and mempool (unconfirmed).
func (s *MultiChainLookupService) GetUTXOEntry(outpoint utxopool.Outpoint) (utxopool.UTXOEntry, error) {
	return s.multiChain.GetMainChain().GetUTXOEntry(outpoint)
}

// ContainsUTXO checks if a UTXO exists on the main chain.
// This checks both chainstate (confirmed) and mempool (unconfirmed).
func (s *MultiChainLookupService) ContainsUTXO(outpoint utxopool.Outpoint) bool {
	return s.multiChain.GetMainChain().ContainsUTXO(outpoint)
}

// GetUTXOsByPubKeyHash returns all UTXOs on the main chain belonging to the given PubKeyHash.
// Results include both confirmed (chainstate) and unconfirmed (mempool) UTXOs.
func (s *MultiChainLookupService) GetUTXOsByPubKeyHash(pubKeyHash transaction.PubKeyHash) ([]transaction.UTXO, error) {
	return s.multiChain.GetMainChain().GetUTXOsByPubKeyHash(pubKeyHash)
}

// =========================================================================
// Side Chain Lookup Methods
// =========================================================================

// GetUTXOAtChainTip retrieves a UTXO entry at the specified chain tip.
// This works for both main chain and side chain tips.
// For side chains, it constructs an ephemeral view using stored deltas.
func (s *MultiChainLookupService) GetUTXOAtChainTip(chainTip common.Hash, outpoint utxopool.Outpoint) (utxopool.UTXOEntry, error) {
	view, err := s.multiChain.CreateViewAtTip(chainTip)
	if err != nil {
		return utxopool.UTXOEntry{}, err
	}
	return view.Get(outpoint)
}

// ContainsUTXOAtChainTip checks if a UTXO exists at the specified chain tip.
// This works for both main chain and side chain tips.
func (s *MultiChainLookupService) ContainsUTXOAtChainTip(chainTip common.Hash, outpoint utxopool.Outpoint) (bool, error) {
	view, err := s.multiChain.CreateViewAtTip(chainTip)
	if err != nil {
		return false, err
	}
	_, err = view.Get(outpoint)
	if err != nil {
		if err == ErrUTXONotFound || err == ErrUTXOAlreadySpent {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// =========================================================================
// Chain Information Methods
// =========================================================================

// GetMainChainTip returns the current main chain tip hash.
func (s *MultiChainLookupService) GetMainChainTip() common.Hash {
	return s.multiChain.GetMainChainTip()
}

// GetSideChainTips returns all known side chain tip hashes.
func (s *MultiChainLookupService) GetSideChainTips() []common.Hash {
	deltas := s.multiChain.GetAllSideChainDeltas()
	tips := make([]common.Hash, len(deltas))
	for i, delta := range deltas {
		tips[i] = delta.ChainTip
	}
	return tips
}

// GetSideChainInfo returns information about a specific side chain.
func (s *MultiChainLookupService) GetSideChainInfo(chainTip common.Hash) (*SideChainInfo, bool) {
	delta, exists := s.multiChain.GetSideChainDelta(chainTip)
	if !exists {
		return nil, false
	}
	return &SideChainInfo{
		ChainTip:   delta.ChainTip,
		ForkPoint:  delta.ForkPoint,
		BlockCount: delta.BlockCount,
		UTXOsAdded: uint64(len(delta.AddedUTXOs)),
		UTXOsSpent: uint64(len(delta.SpentUTXOs)),
	}, true
}

// SideChainInfo contains summary information about a side chain.
type SideChainInfo struct {
	ChainTip   common.Hash
	ForkPoint  common.Hash
	BlockCount uint64
	UTXOsAdded uint64
	UTXOsSpent uint64
}

// CanLookupAtChainTip checks if a UTXO lookup can be performed at the given chain tip.
func (s *MultiChainLookupService) CanLookupAtChainTip(chainTip common.Hash) bool {
	return s.multiChain.CanCreateViewAt(chainTip)
}

// =========================================================================
// Access to Underlying Services
// =========================================================================

// GetMultiChainService returns the underlying MultiChainUTXOService.
// Use this for block validation and other advanced operations.
func (s *MultiChainLookupService) GetMultiChainService() *MultiChainUTXOService {
	return s.multiChain
}

// GetMainChainService returns the main chain FullNodeUTXOService.
// Use this for main chain operations that need direct access.
func (s *MultiChainLookupService) GetMainChainService() *FullNodeUTXOService {
	return s.multiChain.GetMainChain()
}
