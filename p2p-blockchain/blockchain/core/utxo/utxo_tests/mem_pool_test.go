package utxo_tests

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/utxo"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/utxopool"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"testing"
)

func TestMemPool_AddAndGet(t *testing.T) {
	pool := utxo.NewMemUTXOPool()

	var txID transaction.TransactionID
	txID[0] = 1
	outpoint := utxopool.NewOutpoint(txID, 0)

	entry := utxopool.NewUTXOEntry(
		transaction.Output{Value: 1000, PubKeyHash: transaction.PubKeyHash{}},
		0, // unconfirmed
		false,
	)

	// Add UTXO
	if err := pool.Add(outpoint, entry); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Get UTXO
	got, err := pool.Get(outpoint)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.Output.Value != 1000 {
		t.Errorf("Value mismatch: got %d, want 1000", got.Output.Value)
	}

	// Check via GetUTXO interface
	output, err := pool.GetUTXO(txID, 0)
	if err != nil {
		t.Fatalf("GetUTXO failed: %v", err)
	}
	if output.Value != 1000 {
		t.Errorf("Value mismatch: got %d, want 1000", output.Value)
	}
}

func TestMemPool_MarkSpent(t *testing.T) {
	pool := utxo.NewMemUTXOPool()

	var txID transaction.TransactionID
	txID[0] = 1
	outpoint := utxopool.NewOutpoint(txID, 0)

	// Mark as spent (simulating a chainstate UTXO being spent)
	pool.MarkSpent(outpoint)

	if !pool.IsSpent(outpoint) {
		t.Error("Expected IsSpent to return true")
	}

	// Unmark
	pool.UnmarkSpent(outpoint)

	if pool.IsSpent(outpoint) {
		t.Error("Expected IsSpent to return false after unmark")
	}
}
