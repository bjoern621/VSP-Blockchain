package utxo_tests

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/utxo"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/utxopool"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockBaseProvider implements BaseUTXOProvider for testing
type mockBaseProvider struct {
	utxos map[string]utxopool.UTXOEntry
}

func newMockBaseProvider() *mockBaseProvider {
	return &mockBaseProvider{
		utxos: make(map[string]utxopool.UTXOEntry),
	}
}

func (m *mockBaseProvider) Get(outpoint utxopool.Outpoint) (utxopool.UTXOEntry, error) {
	key := string(outpoint.Key())
	if entry, ok := m.utxos[key]; ok {
		return entry, nil
	}
	return utxopool.UTXOEntry{}, utxo.ErrUTXONotFound
}

func (m *mockBaseProvider) Contains(outpoint utxopool.Outpoint) bool {
	key := string(outpoint.Key())
	_, ok := m.utxos[key]
	return ok
}

func (m *mockBaseProvider) Add(outpoint utxopool.Outpoint, entry utxopool.UTXOEntry) {
	m.utxos[string(outpoint.Key())] = entry
}

// createTestTxID creates a deterministic transaction ID for testing
func createTestTxID(seed byte) transaction.TransactionID {
	var txID transaction.TransactionID
	for i := range txID {
		txID[i] = seed
	}
	return txID
}

// createTestOutput creates a test output with given value
func createTestOutput(value uint64) transaction.Output {
	var pubKeyHash transaction.PubKeyHash
	return transaction.Output{
		Value:      value,
		PubKeyHash: pubKeyHash,
	}
}

// TestEphemeralUTXOView_GetFromBase tests that view correctly delegates to base
func TestEphemeralUTXOView_GetFromBase(t *testing.T) {
	base := newMockBaseProvider()
	txID := createTestTxID(1)
	outpoint := utxopool.NewOutpoint(txID, 0)
	entry := utxopool.NewUTXOEntry(createTestOutput(1000), 10, false)
	base.Add(outpoint, entry)

	view := utxo.NewEphemeralUTXOView(base)

	// Should find the UTXO from base
	got, err := view.Get(outpoint)
	require.NoError(t, err)
	assert.Equal(t, entry.Output.Value, got.Output.Value)
}

// TestEphemeralUTXOView_GetFromOverlay tests that view returns overlayed UTXOs
func TestEphemeralUTXOView_GetFromOverlay(t *testing.T) {
	base := newMockBaseProvider()
	view := utxo.NewEphemeralUTXOView(base)

	// Create a transaction that creates a new UTXO
	tx := &transaction.Transaction{
		Inputs:  []transaction.Input{},
		Outputs: []transaction.Output{createTestOutput(500)},
	}
	txID := createTestTxID(2)

	err := view.ApplyTx(tx, txID, 10, true)
	require.NoError(t, err)

	// Should find the UTXO from overlay
	outpoint := utxopool.NewOutpoint(txID, 0)
	got, err := view.Get(outpoint)
	require.NoError(t, err)
	assert.Equal(t, uint64(500), got.Output.Value)
}

// TestEphemeralUTXOView_SpendHidesUTXO tests that spent UTXOs are hidden
func TestEphemeralUTXOView_SpendHidesUTXO(t *testing.T) {
	base := newMockBaseProvider()
	txID1 := createTestTxID(1)
	outpoint := utxopool.NewOutpoint(txID1, 0)
	entry := utxopool.NewUTXOEntry(createTestOutput(1000), 10, false)
	base.Add(outpoint, entry)

	view := utxo.NewEphemeralUTXOView(base)

	// Create a transaction that spends the UTXO
	tx := &transaction.Transaction{
		Inputs: []transaction.Input{
			{PrevTxID: txID1, OutputIndex: 0},
		},
		Outputs: []transaction.Output{createTestOutput(900)},
	}
	txID2 := createTestTxID(2)

	err := view.ApplyTx(tx, txID2, 11, false)
	require.NoError(t, err)

	// Original UTXO should now be spent
	assert.True(t, view.IsSpent(outpoint))

	// Trying to get spent UTXO should return error
	_, err = view.Get(outpoint)
	assert.ErrorIs(t, err, utxo.ErrUTXOAlreadySpent)
}

// TestEphemeralUTXOView_DoubleSpendPrevented tests that double spends are rejected
func TestEphemeralUTXOView_DoubleSpendPrevented(t *testing.T) {
	base := newMockBaseProvider()
	txID1 := createTestTxID(1)
	outpoint := utxopool.NewOutpoint(txID1, 0)
	entry := utxopool.NewUTXOEntry(createTestOutput(1000), 10, false)
	base.Add(outpoint, entry)

	view := utxo.NewEphemeralUTXOView(base)

	// First spend should succeed
	tx1 := &transaction.Transaction{
		Inputs:  []transaction.Input{{PrevTxID: txID1, OutputIndex: 0}},
		Outputs: []transaction.Output{createTestOutput(500)},
	}
	err := view.ApplyTx(tx1, createTestTxID(2), 11, false)
	require.NoError(t, err)

	// Second spend of same UTXO should fail
	tx2 := &transaction.Transaction{
		Inputs:  []transaction.Input{{PrevTxID: txID1, OutputIndex: 0}},
		Outputs: []transaction.Output{createTestOutput(500)},
	}
	err = view.ApplyTx(tx2, createTestTxID(3), 11, false)
	assert.ErrorIs(t, err, utxo.ErrUTXOAlreadySpent)
}

// TestEphemeralUTXOView_SpendNonExistentFails tests spending non-existent UTXO
func TestEphemeralUTXOView_SpendNonExistentFails(t *testing.T) {
	base := newMockBaseProvider()
	view := utxo.NewEphemeralUTXOView(base)

	tx := &transaction.Transaction{
		Inputs:  []transaction.Input{{PrevTxID: createTestTxID(99), OutputIndex: 0}},
		Outputs: []transaction.Output{createTestOutput(500)},
	}
	err := view.ApplyTx(tx, createTestTxID(2), 11, false)
	assert.ErrorIs(t, err, utxo.ErrUTXONotFound)
}

// TestEphemeralUTXOView_GetAddedUTXOs tests delta extraction
func TestEphemeralUTXOView_GetAddedUTXOs(t *testing.T) {
	base := newMockBaseProvider()
	view := utxo.NewEphemeralUTXOView(base)

	// Apply a coinbase transaction
	tx := &transaction.Transaction{
		Inputs:  []transaction.Input{},
		Outputs: []transaction.Output{createTestOutput(5000), createTestOutput(1000)},
	}
	txID := createTestTxID(1)
	err := view.ApplyTx(tx, txID, 10, true)
	require.NoError(t, err)

	added := view.GetAddedUTXOs()
	assert.Len(t, added, 2)
}

// TestEphemeralUTXOView_GetSpentOutpoints tests delta extraction
func TestEphemeralUTXOView_GetSpentOutpoints(t *testing.T) {
	base := newMockBaseProvider()
	txID1 := createTestTxID(1)
	base.Add(utxopool.NewOutpoint(txID1, 0), utxopool.NewUTXOEntry(createTestOutput(1000), 10, false))
	base.Add(utxopool.NewOutpoint(txID1, 1), utxopool.NewUTXOEntry(createTestOutput(2000), 10, false))

	view := utxo.NewEphemeralUTXOView(base)

	// Spend both UTXOs
	tx := &transaction.Transaction{
		Inputs: []transaction.Input{
			{PrevTxID: txID1, OutputIndex: 0},
			{PrevTxID: txID1, OutputIndex: 1},
		},
		Outputs: []transaction.Output{createTestOutput(2500)},
	}
	err := view.ApplyTx(tx, createTestTxID(2), 11, false)
	require.NoError(t, err)

	spent := view.GetSpentOutpoints()
	assert.Len(t, spent, 2)
}

// TestEphemeralUTXOView_ChainedTransactions tests spending outputs created in same view
func TestEphemeralUTXOView_ChainedTransactions(t *testing.T) {
	base := newMockBaseProvider()
	view := utxo.NewEphemeralUTXOView(base)

	// Tx1: Coinbase creates UTXO
	tx1 := &transaction.Transaction{
		Inputs:  []transaction.Input{},
		Outputs: []transaction.Output{createTestOutput(5000)},
	}
	txID1 := createTestTxID(1)
	err := view.ApplyTx(tx1, txID1, 10, true)
	require.NoError(t, err)

	// Tx2: Spend the coinbase output (created in same view)
	tx2 := &transaction.Transaction{
		Inputs:  []transaction.Input{{PrevTxID: txID1, OutputIndex: 0}},
		Outputs: []transaction.Output{createTestOutput(4000), createTestOutput(900)},
	}
	txID2 := createTestTxID(2)
	err = view.ApplyTx(tx2, txID2, 10, false)
	require.NoError(t, err)

	// The coinbase output should be gone from added (it was spent)
	added := view.GetAddedUTXOs()
	assert.Len(t, added, 2) // Only tx2's outputs remain

	// It should not appear in spent (it was created and consumed in same view)
	spent := view.GetSpentOutpoints()
	assert.Len(t, spent, 0)
}

// TestEphemeralUTXOView_DoesNotMutateBase tests that base is not modified
func TestEphemeralUTXOView_DoesNotMutateBase(t *testing.T) {
	base := newMockBaseProvider()
	txID1 := createTestTxID(1)
	outpoint := utxopool.NewOutpoint(txID1, 0)
	entry := utxopool.NewUTXOEntry(createTestOutput(1000), 10, false)
	base.Add(outpoint, entry)

	view := utxo.NewEphemeralUTXOView(base)

	// Spend the UTXO in the view
	tx := &transaction.Transaction{
		Inputs:  []transaction.Input{{PrevTxID: txID1, OutputIndex: 0}},
		Outputs: []transaction.Output{createTestOutput(900)},
	}
	err := view.ApplyTx(tx, createTestTxID(2), 11, false)
	require.NoError(t, err)

	// Base should still have the UTXO
	assert.True(t, base.Contains(outpoint))
	got, err := base.Get(outpoint)
	require.NoError(t, err)
	assert.Equal(t, uint64(1000), got.Output.Value)
}
