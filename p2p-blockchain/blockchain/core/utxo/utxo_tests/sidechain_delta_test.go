package utxo_tests

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/utxo"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/utxopool"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestHash creates a deterministic hash for testing
func createTestHash(seed byte) common.Hash {
	var hash common.Hash
	for i := range hash {
		hash[i] = seed
	}
	return hash
}

// TestSideChainDelta_NewDelta tests creation of new delta
func TestSideChainDelta_NewDelta(t *testing.T) {
	forkPoint := createTestHash(1)
	chainTip := createTestHash(2)

	delta := utxo.NewSideChainDelta(forkPoint, chainTip)

	assert.Equal(t, forkPoint, delta.ForkPoint)
	assert.Equal(t, chainTip, delta.ChainTip)
	assert.Empty(t, delta.AddedUTXOs)
	assert.Empty(t, delta.SpentUTXOs)
	assert.Equal(t, uint64(0), delta.BlockCount)
}

// TestSideChainDelta_ApplyView tests merging view changes into delta
func TestSideChainDelta_ApplyView(t *testing.T) {
	forkPoint := createTestHash(1)
	chainTip := createTestHash(2)
	newTip := createTestHash(3)

	delta := utxo.NewSideChainDelta(forkPoint, chainTip)

	// Create a mock view with some changes
	base := newMockBaseProvider()
	txID1 := createTestTxID(1)
	base.Add(utxopool.NewOutpoint(txID1, 0), utxopool.NewUTXOEntry(createTestOutput(1000), 10, false))

	view := utxo.NewEphemeralUTXOView(base)

	// Spend existing UTXO and create new one
	tx := &transaction.Transaction{
		Inputs:  []transaction.Input{{PrevTxID: txID1, OutputIndex: 0}},
		Outputs: []transaction.Output{createTestOutput(900)},
	}
	txID2 := createTestTxID(2)
	err := view.ApplyTx(tx, txID2, 11, false)
	require.NoError(t, err)

	// Apply view to delta
	delta.ApplyView(view, newTip)

	assert.Equal(t, newTip, delta.ChainTip)
	assert.Equal(t, uint64(1), delta.BlockCount)
	assert.Len(t, delta.AddedUTXOs, 1)
	assert.Len(t, delta.SpentUTXOs, 1)
}

// TestSideChainDelta_Clone tests deep copying
func TestSideChainDelta_Clone(t *testing.T) {
	forkPoint := createTestHash(1)
	chainTip := createTestHash(2)

	original := utxo.NewSideChainDelta(forkPoint, chainTip)

	// Add some data
	txID := createTestTxID(1)
	outpoint := utxopool.NewOutpoint(txID, 0)
	original.AddedUTXOs[string(outpoint.Key())] = utxopool.NewUTXOEntry(createTestOutput(1000), 10, false)
	original.SpentUTXOs[string(utxopool.NewOutpoint(txID, 1).Key())] = struct{}{}
	original.BlockCount = 5

	// Clone and verify independence
	cloned := original.Clone()

	assert.Equal(t, original.ForkPoint, cloned.ForkPoint)
	assert.Equal(t, original.ChainTip, cloned.ChainTip)
	assert.Equal(t, original.BlockCount, cloned.BlockCount)
	assert.Len(t, cloned.AddedUTXOs, 1)
	assert.Len(t, cloned.SpentUTXOs, 1)

	// Modify original, clone should not change
	original.BlockCount = 999
	original.AddedUTXOs["new"] = utxopool.UTXOEntry{}
	assert.Equal(t, uint64(5), cloned.BlockCount)
	assert.Len(t, cloned.AddedUTXOs, 1)
}

// TestSideChainDelta_NetEffectOnChainedSpend tests that spending a UTXO
// added in the same delta results in net zero
func TestSideChainDelta_NetEffectOnChainedSpend(t *testing.T) {
	forkPoint := createTestHash(1)
	chainTip := createTestHash(2)
	delta := utxo.NewSideChainDelta(forkPoint, chainTip)

	base := newMockBaseProvider()
	view := utxo.NewEphemeralUTXOView(base)

	// Block 1: Create a UTXO
	tx1 := &transaction.Transaction{
		Inputs:  []transaction.Input{},
		Outputs: []transaction.Output{createTestOutput(5000)},
	}
	txID1 := createTestTxID(1)
	err := view.ApplyTx(tx1, txID1, 10, true)
	require.NoError(t, err)
	delta.ApplyView(view, createTestHash(3))

	// Block 2: Spend the UTXO created in block 1
	deltaProvider := utxo.NewDeltaBaseProvider(base, delta)
	view2 := utxo.NewEphemeralUTXOView(deltaProvider)

	tx2 := &transaction.Transaction{
		Inputs:  []transaction.Input{{PrevTxID: txID1, OutputIndex: 0}},
		Outputs: []transaction.Output{createTestOutput(4000)},
	}
	txID2 := createTestTxID(2)
	err = view2.ApplyTx(tx2, txID2, 11, false)
	require.NoError(t, err)
	delta.ApplyView(view2, createTestHash(4))

	// The original UTXO was added then removed - should not be in added
	outpoint1 := utxopool.NewOutpoint(txID1, 0)
	_, exists := delta.AddedUTXOs[string(outpoint1.Key())]
	assert.False(t, exists)

	// Should not be marked as spent from main chain either (it never was on main chain)
	_, isSpent := delta.SpentUTXOs[string(outpoint1.Key())]
	assert.False(t, isSpent)

	// Only tx2's output should remain
	assert.Len(t, delta.AddedUTXOs, 1)
}

// TestSideChainDeltaStore_BasicOperations tests store CRUD operations
func TestSideChainDeltaStore_BasicOperations(t *testing.T) {
	store := utxo.NewSideChainDeltaStore()

	forkPoint := createTestHash(1)
	chainTip := createTestHash(2)
	delta := utxo.NewSideChainDelta(forkPoint, chainTip)

	// Put
	store.Put(delta)
	assert.Equal(t, 1, store.Size())

	// Get
	got, ok := store.Get(chainTip)
	assert.True(t, ok)
	assert.Equal(t, forkPoint, got.ForkPoint)

	// Get non-existent
	_, ok = store.Get(createTestHash(99))
	assert.False(t, ok)

	// GetByForkPoint
	tips := store.GetByForkPoint(forkPoint)
	assert.Len(t, tips, 1)
	assert.Equal(t, chainTip, tips[0])

	// Remove
	store.Remove(chainTip)
	assert.Equal(t, 0, store.Size())
	_, ok = store.Get(chainTip)
	assert.False(t, ok)
}

// TestSideChainDeltaStore_MultipleForks tests multiple side chains from same fork
func TestSideChainDeltaStore_MultipleForks(t *testing.T) {
	store := utxo.NewSideChainDeltaStore()

	forkPoint := createTestHash(1)
	tip1 := createTestHash(2)
	tip2 := createTestHash(3)

	delta1 := utxo.NewSideChainDelta(forkPoint, tip1)
	delta2 := utxo.NewSideChainDelta(forkPoint, tip2)

	store.Put(delta1)
	store.Put(delta2)

	assert.Equal(t, 2, store.Size())

	tips := store.GetByForkPoint(forkPoint)
	assert.Len(t, tips, 2)
}

// TestSideChainDeltaStore_PruneBefore tests pruning old deltas
func TestSideChainDeltaStore_PruneBefore(t *testing.T) {
	store := utxo.NewSideChainDeltaStore()

	ancientFork := createTestHash(1)
	recentFork := createTestHash(2)

	delta1 := utxo.NewSideChainDelta(ancientFork, createTestHash(3))
	delta2 := utxo.NewSideChainDelta(ancientFork, createTestHash(4))
	delta3 := utxo.NewSideChainDelta(recentFork, createTestHash(5))

	store.Put(delta1)
	store.Put(delta2)
	store.Put(delta3)

	assert.Equal(t, 3, store.Size())

	// Prune ancient fork point
	ancientForkPoints := map[common.Hash]struct{}{ancientFork: {}}
	pruned := store.PruneBefore(ancientForkPoints)

	assert.Equal(t, 2, pruned)
	assert.Equal(t, 1, store.Size())

	// Recent fork should still exist
	_, ok := store.Get(createTestHash(5))
	assert.True(t, ok)
}

// TestDeltaBaseProvider tests the delta base provider wrapper
func TestDeltaBaseProvider(t *testing.T) {
	base := newMockBaseProvider()
	txID1 := createTestTxID(1)
	outpoint1 := utxopool.NewOutpoint(txID1, 0)
	outpoint2 := utxopool.NewOutpoint(txID1, 1)
	base.Add(outpoint1, utxopool.NewUTXOEntry(createTestOutput(1000), 10, false))
	base.Add(outpoint2, utxopool.NewUTXOEntry(createTestOutput(2000), 10, false))

	delta := utxo.NewSideChainDelta(createTestHash(1), createTestHash(2))

	// Add a new UTXO in delta
	txID2 := createTestTxID(2)
	outpoint3 := utxopool.NewOutpoint(txID2, 0)
	delta.AddedUTXOs[string(outpoint3.Key())] = utxopool.NewUTXOEntry(createTestOutput(500), 11, false)

	// Mark outpoint1 as spent in delta
	delta.SpentUTXOs[string(outpoint1.Key())] = struct{}{}

	provider := utxo.NewDeltaBaseProvider(base, delta)

	// outpoint1 should be spent
	assert.False(t, provider.Contains(outpoint1))
	_, err := provider.Get(outpoint1)
	assert.ErrorIs(t, err, utxo.ErrUTXOAlreadySpent)

	// outpoint2 should still be accessible (unspent in delta)
	assert.True(t, provider.Contains(outpoint2))
	entry, err := provider.Get(outpoint2)
	require.NoError(t, err)
	assert.Equal(t, uint64(2000), entry.Output.Value)

	// outpoint3 should be accessible (added in delta)
	assert.True(t, provider.Contains(outpoint3))
	entry, err = provider.Get(outpoint3)
	require.NoError(t, err)
	assert.Equal(t, uint64(500), entry.Output.Value)
}
