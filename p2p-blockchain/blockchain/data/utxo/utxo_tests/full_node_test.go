package utxo_tests

import (
	"errors"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/utxo"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/utxopool"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/infrastructure"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"testing"
)

func TestFullNode_LookupOrder(t *testing.T) {
	mempool := utxo.NewMemUTXOPool()
	dao, _ := infrastructure.NewUTXOEntryDAO(infrastructure.UTXOEntryDAOConfig{InMemory: true})
	chainstate, err := utxo.NewChainStateService(utxo.ChainStateConfig{CacheSize: 100}, dao)
	if err != nil {
		t.Fatalf("Failed to create chainstate: %v", err)
	}
	defer func(chainstate *utxo.ChainStateService) {
		err2 := chainstate.Close()
		if err2 != nil {
			t.Fatalf("Failed to close chainstate: %v", err2)
		}
	}(chainstate)

	pool := utxo.NewCombinedUTXOPool(mempool, chainstate)

	var txID1 transaction.TransactionID
	txID1[0] = 1
	outpoint1 := utxopool.NewOutpoint(txID1, 0)

	var txID2 transaction.TransactionID
	txID2[0] = 2
	outpoint2 := utxopool.NewOutpoint(txID2, 0)

	// Add to chainstate
	chainstateEntry := utxopool.NewUTXOEntry(
		transaction.Output{Value: 1000},
		100,
		false,
	)
	if err = chainstate.Add(outpoint1, chainstateEntry); err != nil {
		t.Fatalf("chainstate.Add failed: %v", err)
	}

	// Add to mempool
	mempoolEntry := utxopool.NewUTXOEntry(
		transaction.Output{Value: 2000},
		0,
		false,
	)
	if err = mempool.Add(outpoint2, mempoolEntry); err != nil {
		t.Fatalf("mempool.Add failed: %v", err)
	}

	// Test: Get chainstate UTXO via combined pool
	var output transaction.Output
	output, err = pool.GetUTXO(txID1, 0)
	if err != nil {
		t.Error("Failed to get chainstate UTXO")
	}
	if output.Value != 1000 {
		t.Errorf("Chainstate value mismatch: got %d, want 1000", output.Value)
	}

	// Test: Get mempool UTXO via combined pool
	output, err = pool.GetUTXO(txID2, 0)
	if err != nil {
		t.Error("Failed to get mempool UTXO")
	}
	if output.Value != 2000 {
		t.Errorf("Mempool value mismatch: got %d, want 2000", output.Value)
	}

	// Test: Mark chainstate UTXO as spent via mempool
	mempool.MarkSpent(outpoint1)

	// Now the chainstate UTXO should not be available
	_, err = pool.GetUTXO(txID1, 0)
	if !errors.Is(err, utxo.ErrUTXOAlreadySpent) {
		t.Error("Should not get spent chainstate UTXO")
	}
}

func TestFullNode_ApplyTransaction(t *testing.T) {
	mempool := utxo.NewMemUTXOPool()
	dao, _ := infrastructure.NewUTXOEntryDAO(infrastructure.UTXOEntryDAOConfig{InMemory: true})
	chainstate, err := utxo.NewChainStateService(utxo.ChainStateConfig{CacheSize: 100}, dao)
	if err != nil {
		t.Fatalf("Failed to create chainstate: %v", err)
	}
	defer func(chainstate *utxo.ChainStateService) {
		err2 := chainstate.Close()
		if err2 != nil {
			t.Fatalf("Failed to close chainstate: %v", err2)
		}
	}(chainstate)

	pool := utxo.NewCombinedUTXOPool(mempool, chainstate)

	// Create a previous UTXO in chainstate
	var prevTxID transaction.TransactionID
	prevTxID[0] = 10
	prevOutpoint := utxopool.NewOutpoint(prevTxID, 0)

	prevEntry := utxopool.NewUTXOEntry(
		transaction.Output{Value: 5000},
		50,
		false,
	)
	if err = chainstate.Add(prevOutpoint, prevEntry); err != nil {
		t.Fatalf("chainstate.Add failed: %v", err)
	}

	// Create a transaction that spends the previous UTXO
	var newTxID transaction.TransactionID
	newTxID[0] = 20

	tx := &transaction.Transaction{
		Inputs: []transaction.Input{
			{PrevTxID: prevTxID, OutputIndex: 0},
		},
		Outputs: []transaction.Output{
			{Value: 4000},
			{Value: 900}, // change (fee would be 100)
		},
	}

	// Apply as unconfirmed transaction
	if err = pool.ApplyTransaction(tx, newTxID, 0, false); err != nil {
		t.Fatalf("ApplyTransaction failed: %v", err)
	}

	// The previous UTXO should now be marked as spent
	_, err = pool.GetUTXO(prevTxID, 0)
	if !errors.Is(err, utxo.ErrUTXOAlreadySpent) {
		t.Error("Previous UTXO should be marked as spent")
	}

	// The new outputs should be available in mempool
	var output transaction.Output
	output, err = pool.GetUTXO(newTxID, 0)
	if err != nil {
		t.Error("New output 0 should be available")
	}
	if output.Value != 4000 {
		t.Errorf("Output 0 value mismatch: got %d, want 4000", output.Value)
	}

	output, err = pool.GetUTXO(newTxID, 1)
	if err != nil {
		t.Error("New output 1 should be available")
	}
	if output.Value != 900 {
		t.Errorf("Output 1 value mismatch: got %d, want 900", output.Value)
	}

	// Now confirm the transaction
	if err = pool.ApplyTransaction(tx, newTxID, 100, false); err != nil {
		t.Fatalf("ApplyTransaction (confirm) failed: %v", err)
	}

	// Outputs should now be in chainstate
	var entry utxopool.UTXOEntry
	entry, err = chainstate.Get(utxopool.NewOutpoint(newTxID, 0))
	if err != nil {
		t.Error("Output 0 should be in chainstate after confirmation")
	}
	if entry.BlockHeight != 100 {
		t.Errorf("BlockHeight mismatch: got %d, want 100", entry.BlockHeight)
	}

	// Previous UTXO should be removed from chainstate
	if chainstate.Contains(prevOutpoint) {
		t.Error("Previous UTXO should be removed from chainstate")
	}
}
