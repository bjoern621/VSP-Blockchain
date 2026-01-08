package utxo_tests

import (
	"errors"
	"os"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/utxo"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/utxopool"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/infrastructure"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"testing"
)

func TestChainState_AddAndGet(t *testing.T) {
	// Create temporary directory for test database
	tmpDir, err := os.MkdirTemp("", "chainstate_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func(path string) {
		err2 := os.RemoveAll(path)
		if err2 != nil {
			t.Fatalf("Failed to remove temp dir: %v", err2)
		}
	}(tmpDir)

	daoConfig := infrastructure.NewUTXOEntryDAOConfig(tmpDir, false)
	var dao *infrastructure.UTXOEntryDAOImpl
	dao, err = infrastructure.NewUTXOEntryDAO(daoConfig)
	if err != nil {
		t.Fatalf("Failed to create dao: %v", err)
	}

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
	if !errors.Is(err, utxo.ErrUTXONotFound) {
		t.Fatalf("Expected Get to return Not found after removal, is %v", err)
	}
}

func TestChainState_InMemory(t *testing.T) {
	daoConfig := infrastructure.UTXOEntryDAOConfig{
		InMemory: true,
	}
	dao, err := infrastructure.NewUTXOEntryDAO(daoConfig)
	if err != nil {
		t.Fatalf("Failed to create dao: %v", err)
	}
	var chainstate *utxo.ChainStateService
	chainstate, err = utxo.NewChainStateService(utxo.ChainStateConfig{CacheSize: 100}, dao)
	if err != nil {
		t.Fatalf("Failed to create in-memory chainstate: %v", err)
	}
	defer func(chainstate *utxo.ChainStateService) {
		err2 := chainstate.Close()
		if err2 != nil {
			t.Fatalf("Failed to close chainstate: %v", err2)
		}
	}(chainstate)

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
