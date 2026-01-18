package core

import (
	"fmt"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"testing"
)

// txIDToBlockHash converts a TransactionID (32 bytes) into the block.Hash type used by Mempool.IsKnownTransaction.
func txIDToBlockHash(id transaction.TransactionID) common.Hash {
	var h common.Hash
	copy(h[:], id[:])
	return h
}

type mockValidator struct{}

func (m *mockValidator) ValidateTransaction(_ transaction.Transaction, _ common.Hash) (bool, error) {
	return true, nil
}

// mockBlockStore is a mock implementation of blockchain.BlockStoreAPI for testing
type mockBlockStore2 struct{}

func (m *mockBlockStore2) AddBlock(_ block.Block) []common.Hash {
	return nil
}

func (m *mockBlockStore2) IsOrphanBlock(_ block.Block) (bool, error) {
	return false, nil
}

func (m *mockBlockStore2) IsPartOfMainChain(_ block.Block) bool {
	return true
}

func (m *mockBlockStore2) GetBlockByHash(_ common.Hash) (block.Block, error) {
	return block.Block{}, fmt.Errorf("block not found")
}

func (m *mockBlockStore2) GetBlocksByHeight(_ uint64) []block.Block {
	return nil
}

func (m *mockBlockStore2) GetCurrentHeight() uint64 {
	return 0
}

func (m *mockBlockStore2) GetMainChainTip() block.Block {
	return block.Block{}
}

func (m *mockBlockStore2) GetAllBlocksWithMetadata() []block.BlockWithMetadata {
	return nil
}

func (m *mockBlockStore2) IsBlockInvalid(_ block.Block) (bool, error) {
	return false, nil
}

func newMockBlockStore() BlockStoreAPI {
	return &mockBlockStore2{}
}

func TestNewMempool_StartsEmpty(t *testing.T) {
	m := NewMempool(&mockValidator{}, newMockBlockStore())

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
	m := NewMempool(&mockValidator{}, newMockBlockStore())

	// Two different transactions (Hash() is expected to reflect the content).
	tx1 := transaction.Transaction{
		Inputs: []transaction.Input{
			{PrevTxID: transaction.TransactionID{1}},
		},
		Outputs: []transaction.Output{
			{Value: 10},
		},
	}
	tx2 := transaction.Transaction{
		Inputs: []transaction.Input{
			{PrevTxID: transaction.TransactionID{2}},
		},
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
	m := NewMempool(&mockValidator{}, newMockBlockStore())

	tx := transaction.Transaction{
		Inputs: []transaction.Input{
			{PrevTxID: transaction.TransactionID{123}},
		},
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
	m := NewMempool(&mockValidator{}, newMockBlockStore())

	// Some arbitrary hash that (with overwhelming probability) doesn't match any tx hash in an empty mempool.
	var unknown common.Hash
	unknown[0] = 0x42
	unknown[31] = 0x99

	if m.IsKnownTransactionHash(unknown) {
		t.Fatalf("expected unknown hash to be reported as not known")
	}
}
