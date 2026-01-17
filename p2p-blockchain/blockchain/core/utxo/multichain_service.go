package utxo

import (
	"errors"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/blockchain"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/utxopool"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"sync"
)

var (
	ErrBlockNotFound         = errors.New("block not found")
	ErrForkPointNotFound     = errors.New("fork point not found")
	ErrInvalidBlockParent    = errors.New("block parent not found")
	ErrDeltaNotFound         = errors.New("side chain delta not found")
	ErrCannotReconstructView = errors.New("cannot reconstruct UTXO view at this depth")
)

// ConfirmationDepth is the number of blocks before UTXOs move from mempool to chainstate.
// Blocks at height 1-5 have UTXOs in mempool, blocks at height 6+ have UTXOs in chainstate.
const ConfirmationDepth = 5

// MultiChainUTXOService manages UTXO state for the main chain and all side chains.
//
// Architecture (Bitcoin-style):
// - Chainstate contains UTXOs from confirmed blocks (height > ConfirmationDepth)
// - Mempool is ONLY for main chain unconfirmed transactions
// - Side chain validation NEVER uses mempool
// - Side chains use chainstate + block replay to reconstruct UTXO state
//
// Core invariants:
// - Chainstate is consensus (persistent, authoritative for confirmed blocks)
// - Mempool is policy (only tracks main chain tip, never used for block validation)
// - UTXO views are ephemeral (discarded after validation)
// - Block validation uses chainstate + delta from replayed blocks
type MultiChainUTXOService struct {
	mu sync.RWMutex

	// mainChain is the authoritative UTXO service for the main chain
	mainChain *FullNodeUTXOService

	// sideChains stores deltas for all known side chains
	sideChains *sideChainDeltaStore

	// blockStore provides access to block data
	blockStore blockchain.BlockStoreAPI

	// mainChainTip is the current main chain tip
	mainChainTip common.Hash
}

// NewMultiChainUTXOService creates a new multi-chain UTXO service.
func NewMultiChainUTXOService(
	mainChain *FullNodeUTXOService,
	blockStore blockchain.BlockStoreAPI,
) *MultiChainUTXOService {
	var mainChainTip common.Hash
	// Get initial tip - may panic if block store is empty, which is a programming error
	tip := blockStore.GetMainChainTip()
	mainChainTip = tip.Hash()

	return &MultiChainUTXOService{
		mainChain:    mainChain,
		sideChains:   NewSideChainDeltaStore(),
		blockStore:   blockStore,
		mainChainTip: mainChainTip,
	}
}

// GetMainChain returns the main chain UTXO service for direct access.
// This should only be used for main chain operations (mempool, queries).
func (m *MultiChainUTXOService) GetMainChain() *FullNodeUTXOService {
	return m.mainChain
}

// GetMainChainTip returns the current main chain tip hash.
func (m *MultiChainUTXOService) GetMainChainTip() common.Hash {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.mainChainTip
}

// SetMainChainTip updates the main chain tip (after reorg or chain extension).
func (m *MultiChainUTXOService) SetMainChainTip(tip common.Hash) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mainChainTip = tip
}

// CreateViewAtTip creates an ephemeral UTXO view at the given chain tip.
//
// For side chains, this uses the Bitcoin approach:
// 1. Start with chainstate (confirmed UTXOs only)
// 2. Build delta by replaying main chain blocks from fork point to current tip
// 3. Apply any stored side chain deltas on top
func (m *MultiChainUTXOService) CreateViewAtTip(chainTip common.Hash) (UTXOView, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.createViewAtTipUnlocked(chainTip)
}

// createViewAtTipUnlocked creates a view without locking. Must be called with m.mu held.
func (m *MultiChainUTXOService) createViewAtTipUnlocked(chainTip common.Hash) (UTXOView, error) {
	// If this is the main chain tip, use chainstate + main chain delta for unconfirmed blocks
	if chainTip == m.mainChainTip {
		return m.createMainChainViewUnlocked()
	}

	// Check if this tip has a side chain delta
	if delta, hasDelta := m.sideChains.Get(chainTip); hasDelta {
		return m.createViewFromDeltaUnlocked(delta)
	}

	// No delta exists yet - create view for new side chain
	return m.createViewForNewSideChainUnlocked(chainTip)
}

// createMainChainViewUnlocked creates a view at the main chain tip.
// Uses chainstate + delta built from replaying unconfirmed main chain blocks.
func (m *MultiChainUTXOService) createMainChainViewUnlocked() (UTXOView, error) {
	baseProvider := NewChainStateBaseProvider(m.mainChain.GetChainState())

	// Find the highest confirmed block (where chainstate reflects state)
	confirmedHeight := m.getConfirmedHeightUnlocked()
	currentHeight := m.blockStore.GetCurrentHeight()

	// If we have unconfirmed blocks, build delta from them
	if currentHeight > confirmedHeight {
		mainDelta, err := m.buildMainChainDeltaUnlocked(confirmedHeight, currentHeight)
		if err != nil {
			return nil, err
		}
		if mainDelta != nil && (len(mainDelta.AddedUTXOs) > 0 || len(mainDelta.SpentUTXOs) > 0) {
			baseProvider = NewDeltaBaseProvider(baseProvider, mainDelta)
		}
	}

	return NewEphemeralUTXOView(baseProvider), nil
}

// createViewFromDeltaUnlocked creates a view using an existing side chain delta.
func (m *MultiChainUTXOService) createViewFromDeltaUnlocked(delta *SideChainDelta) (UTXOView, error) {
	baseProvider, err := m.createBaseProviderForDeltaUnlocked(delta)
	if err != nil {
		return nil, err
	}
	return NewEphemeralUTXOView(NewDeltaBaseProvider(baseProvider, delta)), nil
}

// createViewForNewSideChainUnlocked creates a view for a chain tip without stored delta.
// This happens when a block arrives that creates a new fork.
func (m *MultiChainUTXOService) createViewForNewSideChainUnlocked(chainTip common.Hash) (UTXOView, error) {
	// Find the fork point by walking back to main chain
	forkPoint, err := m.findForkPointUnlocked(chainTip)
	if err != nil {
		return nil, err
	}

	// Build base provider at fork point
	baseProvider, err := m.buildBaseProviderAtForkPointUnlocked(forkPoint)
	if err != nil {
		return nil, err
	}

	return NewEphemeralUTXOView(baseProvider), nil
}

// createBaseProviderForDeltaUnlocked creates the base provider for applying a delta.
// For forks from main chain: chainstate + main chain blocks from confirmed height to fork point
// For forks from side chain: recursively build parent chain's provider
func (m *MultiChainUTXOService) createBaseProviderForDeltaUnlocked(delta *SideChainDelta) (BaseUTXOProvider, error) {
	// Check if the delta's fork point is on the main chain
	forkPointBlock, err := m.blockStore.GetBlockByHash(delta.ForkPoint)
	if err != nil {
		return nil, ErrForkPointNotFound
	}

	if m.blockStore.IsPartOfMainChain(forkPointBlock) {
		// Fork is from main chain - build base at fork point
		return m.buildBaseProviderAtForkPointUnlocked(delta.ForkPoint)
	}

	// Fork point is NOT on main chain - it's a branch off a side chain
	// We need to find the delta for the fork point and chain them
	parentDelta, exists := m.sideChains.Get(delta.ForkPoint)
	if !exists {
		// Try to collect a chain of deltas back to main chain
		deltas := m.collectDeltaChainUnlocked(delta.ForkPoint)
		if len(deltas) == 0 {
			return nil, ErrCannotReconstructView
		}
		// Find the base fork point (first delta's fork point should be on main chain)
		baseForkPoint := deltas[0].ForkPoint
		baseProvider, err := m.buildBaseProviderAtForkPointUnlocked(baseForkPoint)
		if err != nil {
			return nil, err
		}
		return NewChainedDeltaBaseProvider(baseProvider, deltas), nil
	}

	// Recursively get the base for the parent delta
	parentBase, err := m.createBaseProviderForDeltaUnlocked(parentDelta)
	if err != nil {
		return nil, err
	}

	return NewDeltaBaseProvider(parentBase, parentDelta), nil
}

// buildBaseProviderAtForkPointUnlocked builds a base provider representing UTXO state at fork point.
// Uses chainstate + delta from replaying main chain blocks between confirmed height and fork point.
func (m *MultiChainUTXOService) buildBaseProviderAtForkPointUnlocked(forkPoint common.Hash) (BaseUTXOProvider, error) {
	baseProvider := NewChainStateBaseProvider(m.mainChain.GetChainState())

	// Get fork point height
	forkPointHeight, err := m.getBlockHeightByHashUnlocked(forkPoint)
	if err != nil {
		return nil, err
	}

	// Get confirmed height (chainstate reflects state up to this height)
	confirmedHeight := m.getConfirmedHeightUnlocked()

	// If fork point is at or below confirmed height, chainstate is sufficient
	if forkPointHeight <= confirmedHeight {
		return baseProvider, nil
	}

	// Need to replay main chain blocks from confirmed height to fork point
	mainDelta, err := m.buildMainChainDeltaRangeUnlocked(confirmedHeight, forkPointHeight)
	if err != nil {
		return nil, err
	}

	if mainDelta != nil && (len(mainDelta.AddedUTXOs) > 0 || len(mainDelta.SpentUTXOs) > 0) {
		return NewDeltaBaseProvider(baseProvider, mainDelta), nil
	}

	return baseProvider, nil
}

// buildMainChainDeltaUnlocked builds a delta from main chain blocks.
// Replays blocks from startHeight+1 to endHeight.
func (m *MultiChainUTXOService) buildMainChainDeltaUnlocked(startHeight, endHeight uint64) (*SideChainDelta, error) {
	return m.buildMainChainDeltaRangeUnlocked(startHeight, endHeight)
}

// buildMainChainDeltaRangeUnlocked builds a delta by replaying main chain blocks in range.
func (m *MultiChainUTXOService) buildMainChainDeltaRangeUnlocked(startHeight, endHeight uint64) (*SideChainDelta, error) {
	if endHeight <= startHeight {
		return nil, nil
	}

	delta := NewSideChainDelta(common.Hash{}, m.mainChainTip)

	for height := startHeight + 1; height <= endHeight; height++ {
		blocks := m.blockStore.GetBlocksByHeight(height)
		for _, blk := range blocks {
			if m.blockStore.IsPartOfMainChain(blk) {
				m.applyBlockToDeltaUnlocked(delta, blk, height)
				break
			}
		}
	}

	return delta, nil
}

// applyBlockToDeltaUnlocked applies a block's transactions to a delta.
func (m *MultiChainUTXOService) applyBlockToDeltaUnlocked(delta *SideChainDelta, blk block.Block, height uint64) {
	for i, tx := range blk.Transactions {
		txID := tx.TransactionId()
		isCoinbase := i == 0 && tx.IsCoinbase()

		// Record spent inputs (skip coinbase)
		if !isCoinbase {
			for _, input := range tx.Inputs {
				outpoint := utxopool.NewOutpoint(input.PrevTxID, input.OutputIndex)
				key := string(outpoint.Key())

				// If this was added in the delta, remove it (net effect: nothing)
				if _, existsInAdded := delta.AddedUTXOs[key]; existsInAdded {
					delete(delta.AddedUTXOs, key)
				} else {
					// Mark as spent from chainstate
					delta.SpentUTXOs[key] = struct{}{}
				}
			}
		}

		// Record created outputs
		for j, output := range tx.Outputs {
			outpoint := utxopool.NewOutpoint(txID, uint32(j))
			key := string(outpoint.Key())
			entry := utxopool.NewUTXOEntry(output, height, isCoinbase && j == 0)
			delta.AddedUTXOs[key] = entry
		}
	}
	delta.BlockCount++
}

// collectDeltaChainUnlocked collects all deltas needed to reconstruct state at a given tip.
// Returns deltas ordered from fork point to tip.
func (m *MultiChainUTXOService) collectDeltaChainUnlocked(targetTip common.Hash) []*SideChainDelta {
	var chain []*SideChainDelta
	currentTip := targetTip

	for {
		delta, exists := m.sideChains.Get(currentTip)
		if !exists {
			break
		}
		chain = append([]*SideChainDelta{delta}, chain...) // prepend

		// Check if fork point is on main chain
		forkBlock, err := m.blockStore.GetBlockByHash(delta.ForkPoint)
		if err != nil {
			break
		}
		if m.blockStore.IsPartOfMainChain(forkBlock) {
			// Reached main chain, done
			break
		}

		// Continue walking back
		currentTip = delta.ForkPoint
	}

	return chain
}

// =========================================================================
// Block Validation
// =========================================================================

// CreateViewForBlockValidation creates a UTXO view for validating a new block.
// The view is based on the block's parent state.
//
// IMPORTANT: This NEVER uses mempool. Block validation is consensus-critical.
func (m *MultiChainUTXOService) CreateViewForBlockValidation(parentHash common.Hash) (UTXOView, error) {
	return m.CreateViewAtTip(parentHash)
}

// ValidateBlock validates a block against its parent chain state.
// This method:
// 1. Creates a UTXO view at the block's parent
// 2. Validates all transactions sequentially
// 3. Returns the view for delta extraction if valid
//
// IMPORTANT: This does NOT mutate any global state.
// IMPORTANT: Mempool is NEVER used for block validation.
func (m *MultiChainUTXOService) ValidateBlock(blk block.Block, blockHeight uint64) (UTXOView, error) {
	// Create view at parent
	view, err := m.CreateViewForBlockValidation(blk.Header.PreviousBlockHash)
	if err != nil {
		return nil, err
	}

	// Validate and apply each transaction to the view
	for i, tx := range blk.Transactions {
		txID := tx.TransactionId()
		isCoinbase := i == 0 && tx.IsCoinbase()

		err := view.ApplyTx(&tx, txID, blockHeight, isCoinbase)
		if err != nil {
			return nil, err
		}
	}

	return view, nil
}

// ValidateAndApplySideChainBlock validates a side chain block and updates its delta.
// Returns the updated or new delta for the side chain.
func (m *MultiChainUTXOService) ValidateAndApplySideChainBlock(blk block.Block, blockHeight uint64) (*SideChainDelta, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	parentHash := blk.Header.PreviousBlockHash
	blockHash := blk.Hash()

	// Get or create delta for the parent
	var delta *SideChainDelta
	var baseProvider BaseUTXOProvider

	if parentDelta, exists := m.sideChains.Get(parentHash); exists {
		// Extend existing side chain - clone the parent delta
		delta = parentDelta.Clone()

		// Get base provider for the parent delta
		var err error
		baseProvider, err = m.createBaseProviderForDeltaUnlocked(parentDelta)
		if err != nil {
			return nil, err
		}
		baseProvider = NewDeltaBaseProvider(baseProvider, delta)
	} else {
		// New side chain - find the fork point
		forkPoint, err := m.findForkPointUnlocked(parentHash)
		if err != nil {
			return nil, err
		}
		delta = NewSideChainDeltaWithParent(forkPoint, parentHash, forkPoint)

		// Build base provider at fork point (using block replay, not mempool)
		baseProvider, err = m.buildBaseProviderAtForkPointUnlocked(forkPoint)
		if err != nil {
			return nil, err
		}
	}

	// Create view on top of base for validation
	view := NewEphemeralUTXOView(baseProvider)

	// Validate and apply each transaction
	for i, tx := range blk.Transactions {
		txID := tx.TransactionId()
		isCoinbase := i == 0 && tx.IsCoinbase()

		err := view.ApplyTx(&tx, txID, blockHeight, isCoinbase)
		if err != nil {
			return nil, err
		}
	}

	// Apply view changes to delta
	delta.ApplyView(view, blockHash)

	// Store the updated delta
	m.sideChains.Put(delta)

	return delta, nil
}

// =========================================================================
// Reorg Handling
// =========================================================================

// PromoteSideChainToMain promotes a side chain to become the main chain.
// The old main chain becomes a side chain, preserving all branch information.
//
// Process:
// 1. Build a delta for the old main chain (preserves it as a side chain)
// 2. Roll back main chain UTXO to fork point (using undo data)
// 3. Apply side chain delta forward into main chain UTXO
// 4. Switch active chain tip
// 5. Clear mempool (caller's responsibility to re-accept transactions)
func (m *MultiChainUTXOService) PromoteSideChainToMain(
	newTip common.Hash,
	forkPoint common.Hash,
	disconnectedBlocks []block.Block,
	connectedBlocks []block.Block,
) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	oldMainTip := m.mainChainTip

	// Step 1: Create a delta for the old main chain (now becomes a side chain)
	if len(disconnectedBlocks) > 0 {
		oldMainDelta := m.buildDeltaForDisconnectedBlocksUnlocked(forkPoint, oldMainTip, disconnectedBlocks)
		m.sideChains.Put(oldMainDelta)
	}

	// Step 2: Revert disconnected blocks from chainstate
	for i := len(disconnectedBlocks) - 1; i >= 0; i-- {
		blk := disconnectedBlocks[i]
		if err := m.revertBlockUnlocked(blk); err != nil {
			return err
		}
	}

	// Step 3: Apply connected blocks to chainstate
	for _, blk := range connectedBlocks {
		if err := m.applyBlockUnlocked(blk); err != nil {
			return err
		}
	}

	// Step 4: Update main chain tip
	m.mainChainTip = newTip

	// Step 5: Remove the promoted side chain delta (it's now main chain)
	m.sideChains.Remove(newTip)

	return nil
}

// buildDeltaForDisconnectedBlocksUnlocked builds a delta representing the old main chain.
// This preserves the old main chain as a side chain after reorg.
func (m *MultiChainUTXOService) buildDeltaForDisconnectedBlocksUnlocked(
	forkPoint, oldMainTip common.Hash,
	disconnectedBlocks []block.Block,
) *SideChainDelta {
	delta := NewSideChainDelta(forkPoint, oldMainTip)

	// Build the delta by walking disconnected blocks (fork point -> old tip)
	for _, blk := range disconnectedBlocks {
		blockHeight, _ := m.getBlockHeightByHashUnlocked(blk.Hash())
		m.applyBlockToDeltaUnlocked(delta, blk, blockHeight)
	}

	return delta
}

// revertBlockUnlocked reverts a block from the main chain UTXO state.
// Must be called with m.mu held.
func (m *MultiChainUTXOService) revertBlockUnlocked(blk block.Block) error {
	// Process transactions in reverse order
	for i := len(blk.Transactions) - 1; i >= 0; i-- {
		tx := blk.Transactions[i]
		txID := tx.TransactionId()

		// Get input UTXOs for restoration (simplified - in production would use undo data)
		inputUTXOs := m.getInputUTXOsForRevert(&tx)

		err := m.mainChain.RevertTransaction(&tx, txID, inputUTXOs)
		if err != nil {
			return err
		}
	}
	return nil
}

// applyBlockUnlocked applies a block to the main chain UTXO state.
// Must be called with m.mu held.
func (m *MultiChainUTXOService) applyBlockUnlocked(blk block.Block) error {
	blockHeight, _ := m.getBlockHeightByHashUnlocked(blk.Hash())

	for i, tx := range blk.Transactions {
		txID := tx.TransactionId()
		isCoinbase := i == 0 && tx.IsCoinbase()

		err := m.mainChain.ApplyTransaction(&tx, txID, blockHeight, isCoinbase)
		if err != nil {
			return err
		}
	}
	return nil
}

// getInputUTXOsForRevert retrieves input UTXOs needed for reverting a transaction.
// In a production system, this would use undo data stored during block connection.
func (m *MultiChainUTXOService) getInputUTXOsForRevert(tx *transaction.Transaction) []utxopool.UTXOEntry {
	if tx.IsCoinbase() {
		return []utxopool.UTXOEntry{}
	}
	// In production: Load from undo data stored when block was connected
	return make([]utxopool.UTXOEntry, len(tx.Inputs))
}

// =========================================================================
// Pruning
// =========================================================================

// PruneSideChains removes side chain deltas that are older than the given depth.
// Returns the number of deltas pruned.
func (m *MultiChainUTXOService) PruneSideChains(maxDepth uint64) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find fork points that are too deep
	ancientForkPoints := make(map[common.Hash]struct{})
	mainTip := m.blockStore.GetMainChainTip()
	mainHeight := m.blockStore.GetCurrentHeight()

	// Walk back from main chain tip to find which fork points are too old
	currentHash := mainTip.Hash()
	depth := uint64(0)

	for depth < mainHeight {
		blk, err := m.blockStore.GetBlockByHash(currentHash)
		if err != nil {
			break
		}

		if depth >= maxDepth {
			ancientForkPoints[currentHash] = struct{}{}
		}

		currentHash = blk.Header.PreviousBlockHash
		depth++
	}

	return m.sideChains.PruneBefore(ancientForkPoints)
}

// =========================================================================
// Helper Methods
// =========================================================================

// findForkPointUnlocked finds the fork point between a side chain block and main chain.
// Must be called with m.mu held for reading.
func (m *MultiChainUTXOService) findForkPointUnlocked(sideChainHash common.Hash) (common.Hash, error) {
	currentHash := sideChainHash

	for {
		blk, err := m.blockStore.GetBlockByHash(currentHash)
		if err != nil {
			return common.Hash{}, ErrBlockNotFound
		}

		if m.blockStore.IsPartOfMainChain(blk) {
			return currentHash, nil
		}

		// Genesis check
		if blk.Header.PreviousBlockHash == (common.Hash{}) {
			return currentHash, nil
		}

		currentHash = blk.Header.PreviousBlockHash
	}
}

// getConfirmedHeightUnlocked returns the height up to which chainstate is current.
// UTXOs from blocks at height > confirmedHeight are still in mempool.
func (m *MultiChainUTXOService) getConfirmedHeightUnlocked() uint64 {
	currentHeight := m.blockStore.GetCurrentHeight()
	if currentHeight <= ConfirmationDepth {
		return 0
	}
	return currentHeight - ConfirmationDepth
}

// getBlockHeightByHashUnlocked gets the height of a block by its hash.
func (m *MultiChainUTXOService) getBlockHeightByHashUnlocked(blockHash common.Hash) (uint64, error) {
	height := uint64(0)
	currentHash := blockHash

	for {
		currentBlock, err := m.blockStore.GetBlockByHash(currentHash)
		if err != nil {
			return 0, err
		}

		if currentBlock.Header.PreviousBlockHash == (common.Hash{}) {
			return height, nil
		}

		height++
		currentHash = currentBlock.Header.PreviousBlockHash
	}
}

// GetSideChainDelta retrieves the delta for a side chain tip.
func (m *MultiChainUTXOService) GetSideChainDelta(chainTip common.Hash) (*SideChainDelta, bool) {
	return m.sideChains.Get(chainTip)
}

// GetSideChainCount returns the number of tracked side chains.
func (m *MultiChainUTXOService) GetSideChainCount() int {
	return m.sideChains.Size()
}

// GetAllSideChainDeltas returns all tracked side chain deltas.
func (m *MultiChainUTXOService) GetAllSideChainDeltas() []*SideChainDelta {
	return m.sideChains.GetAll()
}

// ClearMempool clears the mempool. Called after reorg.
func (m *MultiChainUTXOService) ClearMempool() {
	m.mainChain.ClearMempool()
}

// CanCreateViewAt checks if a UTXO view can be created at the given chain tip.
func (m *MultiChainUTXOService) CanCreateViewAt(chainTip common.Hash) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Main chain tip is always available
	if chainTip == m.mainChainTip {
		return true
	}

	// Check if we have a delta for this tip
	if _, exists := m.sideChains.Get(chainTip); exists {
		return true
	}

	// Check if we can find a path to main chain
	_, err := m.findForkPointUnlocked(chainTip)
	return err == nil
}
