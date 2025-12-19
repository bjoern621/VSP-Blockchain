package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"testing"

	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
)

// txIDToBlockHash converts a TransactionID (32 bytes) into the block.Hash type used by Mempool.IsKnownTransaction.
func txIDToBlockHash(id transaction.TransactionID) common.Hash {
	var h common.Hash
	copy(h[:], id[:])
	return h
}

func TestNewMempool_StartsEmpty(t *testing.T) {
	m := NewMempool()

	if m == nil {
		t.Fatalf("expected mempool to be non-nil")
	}
	if m.transactions == nil {
		t.Fatalf("expected transactions map to be initialized")
	}
	if got := len(m.transactions); got != 0 {
		t.Fatalf("expected empty mempool, got %d entries", got)
	}
}

func TestMempool_AddTransaction_MakesTransactionKnownByHash(t *testing.T) {
	m := NewMempool()

	// Two different transactions (Hash() is expected to reflect the content).
	tx1 := transaction.Transaction{
		LockTime: 1,
		Outputs: []transaction.Output{
			{Value: 10},
		},
	}
	tx2 := transaction.Transaction{
		LockTime: 2,
		Outputs: []transaction.Output{
			{Value: 10},
		},
	}

	h1 := txIDToBlockHash(tx1.TransactionId())
	h2 := txIDToBlockHash(tx2.TransactionId())

	// Sanity check for a "reasonable" Hash(): different tx -> different hash.
	// If this fails later, your hashing function likely doesn't include all relevant fields.
	if h1 == h2 {
		t.Fatalf("expected different transactions to have different hashes; got equal hashes %v", h1)
	}

	if m.IsKnownTransactionHash(h1) {
		t.Fatalf("expected tx1 to be unknown before adding it")
	}
	if m.IsKnownTransactionHash(h2) {
		t.Fatalf("expected tx2 to be unknown before adding it")
	}

	m.AddTransaction(tx1)

	if !m.IsKnownTransactionHash(h1) {
		t.Fatalf("expected tx1 to be known after adding it")
	}
	if m.IsKnownTransactionHash(h2) {
		t.Fatalf("expected tx2 to remain unknown when only tx1 was added")
	}

	m.AddTransaction(tx2)

	if !m.IsKnownTransactionHash(h2) {
		t.Fatalf("expected tx2 to be known after adding it")
	}
}

func TestMempool_AddTransaction_DoesNotDuplicateSameTransactionID(t *testing.T) {
	m := NewMempool()

	tx := transaction.Transaction{
		LockTime: 123,
		Outputs: []transaction.Output{
			{Value: 99},
		},
	}
	h := txIDToBlockHash(tx.TransactionId())

	m.AddTransaction(tx)
	if got := len(m.transactions); got != 1 {
		t.Fatalf("expected mempool size 1 after first add, got %d", got)
	}
	if !m.IsKnownTransactionHash(h) {
		t.Fatalf("expected tx to be known after first add")
	}

	// Add the *same* transaction again: should not increase map size.
	m.AddTransaction(tx)
	if got := len(m.transactions); got != 1 {
		t.Fatalf("expected mempool size to remain 1 after duplicate add, got %d", got)
	}
}

func TestMempool_IsKnownTransaction_UnknownHashReturnsFalse(t *testing.T) {
	m := NewMempool()

	// Some arbitrary hash that (with overwhelming probability) doesn't match any tx hash in an empty mempool.
	var unknown common.Hash
	unknown[0] = 0x42
	unknown[31] = 0x99

	if m.IsKnownTransactionHash(unknown) {
		t.Fatalf("expected unknown hash to be reported as not known")
	}
}
