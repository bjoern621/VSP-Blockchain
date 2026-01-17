package utxo_tests

import (
	"errors"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/utxo"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/utxopool"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/infrastructure"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"testing"
)

func TestFullNode_LookupOrder(t *testing.T) {
	mempool := utxo.NewMemUTXOPoolService()
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

	pool := utxo.NewFullNodeUTXOService(mempool, chainstate)

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
	mempool := utxo.NewMemUTXOPoolService()
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

	pool := utxo.NewFullNodeUTXOService(mempool, chainstate)

	// Create a previous UTXO in chainstate (simulating a confirmed UTXO)
	var prevTxID transaction.TransactionID
	prevTxID[0] = 10
	prevOutpoint := utxopool.NewOutpoint(prevTxID, 0)

	prevEntry := utxopool.NewUTXOEntry(
		transaction.Output{Value: 5000},
		50, // Block height 50
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

	// Test 1: Apply as pending mempool transaction (blockHeight=0)
	// This should mark the input as spent (not remove it)
	if err = pool.ApplyTransaction(tx, newTxID, 0, 51, false); err != nil {
		t.Fatalf("ApplyTransaction (mempool) failed: %v", err)
	}

	// The previous UTXO should now be marked as spent
	_, err = pool.GetUTXO(prevTxID, 0)
	if !errors.Is(err, utxo.ErrUTXOAlreadySpent) {
		t.Errorf("Previous UTXO should be marked as spent, got error: %v", err)
	}

	// The UTXO should still exist in chainstate (just marked spent in mempool)
	if !chainstate.Contains(prevOutpoint) {
		t.Error("Previous UTXO should still exist in chainstate when only marked spent")
	}

	// The new outputs should be in mempool (blockHeight=0)
	var output transaction.Output
	output, err = pool.GetUTXO(newTxID, 0)
	if err != nil {
		t.Errorf("New output 0 should be available, got error: %v", err)
	}
	if output.Value != 4000 {
		t.Errorf("Output 0 value mismatch: got %d, want 4000", output.Value)
	}

	output, err = pool.GetUTXO(newTxID, 1)
	if err != nil {
		t.Errorf("New output 1 should be available, got error: %v", err)
	}
	if output.Value != 900 {
		t.Errorf("Output 1 value mismatch: got %d, want 900", output.Value)
	}

	// Verify outputs are in mempool, not chainstate
	if chainstate.Contains(utxopool.NewOutpoint(newTxID, 0)) {
		t.Error("Output 0 should be in mempool, not chainstate for pending tx")
	}

	// Test 2: Now simulate the transaction being included in a block (blockHeight=51)
	// First, clear the mempool state from test 1
	mempool.UnmarkSpent(prevOutpoint)
	_ = mempool.Remove(utxopool.NewOutpoint(newTxID, 0))
	_ = mempool.Remove(utxopool.NewOutpoint(newTxID, 1))

	// Apply as block transaction
	if err = pool.ApplyTransaction(tx, newTxID, 51, 51, false); err != nil {
		t.Fatalf("ApplyTransaction (block) failed: %v", err)
	}

	// Previous UTXO should be removed from chainstate (not just marked spent)
	if chainstate.Contains(prevOutpoint) {
		t.Error("Previous UTXO should be removed from chainstate after block inclusion")
	}

	// New outputs should be in chainstate
	var entry utxopool.UTXOEntry
	entry, err = chainstate.Get(utxopool.NewOutpoint(newTxID, 0))
	if err != nil {
		t.Errorf("Output 0 should be in chainstate after block inclusion, got error: %v", err)
	}
	if entry.BlockHeight != 51 {
		t.Errorf("BlockHeight mismatch: got %d, want 51", entry.BlockHeight)
	}

	entry, err = chainstate.Get(utxopool.NewOutpoint(newTxID, 1))
	if err != nil {
		t.Errorf("Output 1 should be in chainstate after block inclusion, got error: %v", err)
	}
	if entry.BlockHeight != 51 {
		t.Errorf("BlockHeight mismatch: got %d, want 51", entry.BlockHeight)
	}
}
