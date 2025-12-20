package utxo

import (
	"errors"
	"os"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/transaction"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/utxopool"
	"testing"
)

func TestChainState_AddAndGet(t *testing.T) {
	// Create temporary directory for test database
	tmpDir, err := os.MkdirTemp("", "chainstate_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	chainstate, err := NewChainState(ChainStateConfig{
		DBPath:    tmpDir,
		CacheSize: 100,
	})
	if err != nil {
		t.Fatalf("Failed to create chainstate: %v", err)
	}
	defer chainstate.Close()

	var txID transaction.TransactionID
	txID[0] = 1
	outpoint := utxopool.NewOutpoint(txID, 0)

	var pubKeyHash transaction.PubKeyHash
	pubKeyHash[0] = 0xAB
	entry := utxopool.NewUTXOEntry(
		transaction.Output{Value: 5000, PubKeyHash: pubKeyHash},
		100, // confirmed at block 100
		true,
	)

	// Add UTXO
	if err = chainstate.Add(outpoint, entry); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Get UTXO
	var got utxopool.UTXOEntry
	got, err = chainstate.Get(outpoint)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.Output.Value != 5000 {
		t.Errorf("Value mismatch: got %d, want 5000", got.Output.Value)
	}
	if got.BlockHeight != 100 {
		t.Errorf("BlockHeight mismatch: got %d, want 100", got.BlockHeight)
	}
	if !got.IsCoinbase {
		t.Error("Expected IsCoinbase to be true")
	}
	if got.Output.PubKeyHash[0] != 0xAB {
		t.Errorf("PubKeyHash mismatch: got %x, want 0xAB", got.Output.PubKeyHash[0])
	}

	// Remove UTXO
	if err = chainstate.Remove(outpoint); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	// Verify removed
	_, err = chainstate.Get(outpoint)
	if !errors.Is(err, ErrUTXONotFound) {
		t.Fatalf("Expected Get to return Not found after removal, is %v", err)
	}
}

func TestChainState_InMemory(t *testing.T) {
	chainstate, err := NewChainState(ChainStateConfig{
		InMemory:  true,
		CacheSize: 100,
	})
	if err != nil {
		t.Fatalf("Failed to create in-memory chainstate: %v", err)
	}
	defer chainstate.Close()

	var txID transaction.TransactionID
	txID[0] = 2
	outpoint := utxopool.NewOutpoint(txID, 1)

	entry := utxopool.NewUTXOEntry(
		transaction.Output{Value: 3000},
		50,
		false,
	)

	if err = chainstate.Add(outpoint, entry); err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	var got utxopool.UTXOEntry
	got, err = chainstate.Get(outpoint)
	if err != nil || got.Output.Value != 3000 {
		t.Error("Failed to get UTXO from in-memory chainstate")
	}
}
