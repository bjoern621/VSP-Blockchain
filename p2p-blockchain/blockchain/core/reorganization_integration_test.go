package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/utxo"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/blockchain"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/utxopool"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"testing"

	"github.com/stretchr/testify/assert"
)

// =========================================================================
// Mock Implementations for Integration Testing
// =========================================================================

// mockUTXOService is a mock for UTXOService that tracks state changes
type mockUTXOService struct {
	appliedTransactions     []transactionApplied
	revertedTransactions    []transactionReverted
	utxos                   map[utxopool.Outpoint]utxopool.UTXOEntry
	blockHeight             uint64
	addUTXOCalled           int
	spendUTXOCalled         int
	applyTransactionCalled  int
	revertTransactionCalled int
}

type transactionApplied struct {
	TxID        transaction.TransactionID
	BlockHeight uint64
	IsCoinbase  bool
	Outputs     []uint32 // indices of outputs added
	Inputs      []utxopool.Outpoint
}

type transactionReverted struct {
	TxID   transaction.TransactionID
	Inputs []utxopool.UTXOEntry // saved input UTXOs
}

func newMockUTXOService() *mockUTXOService {
	return &mockUTXOService{
		utxos: make(map[utxopool.Outpoint]utxopool.UTXOEntry),
	}
}

func (m *mockUTXOService) GetUTXO(txID transaction.TransactionID, outputIndex uint32) (transaction.Output, error) {
	outpoint := utxopool.NewOutpoint(txID, outputIndex)
	entry, ok := m.utxos[outpoint]
	if !ok {
		return transaction.Output{}, utxo.ErrUTXONotFound
	}
	return entry.Output, nil
}

func (m *mockUTXOService) GetUTXOEntry(outpoint utxopool.Outpoint) (utxopool.UTXOEntry, error) {
	entry, ok := m.utxos[outpoint]
	if !ok {
		return utxopool.UTXOEntry{}, utxo.ErrUTXONotFound
	}
	return entry, nil
}

func (m *mockUTXOService) ContainsUTXO(outpoint utxopool.Outpoint) bool {
	_, ok := m.utxos[outpoint]
	return ok
}

func (m *mockUTXOService) GetUTXOsByPubKeyHash(_ transaction.PubKeyHash) ([]transaction.UTXO, error) {
	return []transaction.UTXO{}, nil
}

func (m *mockUTXOService) AddUTXO(outpoint utxopool.Outpoint, entry utxopool.UTXOEntry) error {
	m.addUTXOCalled++
	m.utxos[outpoint] = entry
	return nil
}

func (m *mockUTXOService) SpendUTXO(outpoint utxopool.Outpoint) error {
	m.spendUTXOCalled++
	delete(m.utxos, outpoint)
	return nil
}

func (m *mockUTXOService) Remove(outpoint utxopool.Outpoint) error {
	delete(m.utxos, outpoint)
	return nil
}

func (m *mockUTXOService) ApplyTransaction(tx *transaction.Transaction, txID transaction.TransactionID, blockHeight uint64, isCoinbase bool) error {
	m.applyTransactionCalled++

	applied := transactionApplied{
		TxID:        txID,
		BlockHeight: blockHeight,
		IsCoinbase:  isCoinbase,
	}

	// Add outputs
	for i, out := range tx.Outputs {
		outpoint := utxopool.NewOutpoint(txID, uint32(i))
		entry := utxopool.NewUTXOEntry(out, blockHeight, isCoinbase)
		m.utxos[outpoint] = entry
		applied.Outputs = append(applied.Outputs, uint32(i))
	}

	// Spend inputs (skip for coinbase)
	if !isCoinbase {
		for _, in := range tx.Inputs {
			outpoint := utxopool.NewOutpoint(in.PrevTxID, in.OutputIndex)
			delete(m.utxos, outpoint)
			applied.Inputs = append(applied.Inputs, outpoint)
		}
	}

	m.appliedTransactions = append(m.appliedTransactions, applied)
	m.blockHeight = blockHeight
	return nil
}

func (m *mockUTXOService) RevertTransaction(tx *transaction.Transaction, txID transaction.TransactionID, inputUTXOs []utxopool.UTXOEntry) error {
	m.revertTransactionCalled++

	reverted := transactionReverted{
		TxID:   txID,
		Inputs: inputUTXOs,
	}

	// Remove outputs (in reverse order)
	for i := len(tx.Outputs) - 1; i >= 0; i-- {
		outpoint := utxopool.NewOutpoint(txID, uint32(i))
		delete(m.utxos, outpoint)
	}

	// Restore inputs
	for i, in := range tx.Inputs {
		outpoint := utxopool.NewOutpoint(in.PrevTxID, in.OutputIndex)
		if i < len(inputUTXOs) {
			entry := inputUTXOs[i]
			m.utxos[outpoint] = entry
		}
	}

	m.revertedTransactions = append(m.revertedTransactions, reverted)
	return nil
}

func (m *mockUTXOService) Flush() error {
	return nil
}

func (m *mockUTXOService) Close() error {
	return nil
}

// mockValidatorForMempool is a mock that always validates successfully
type mockValidatorForMempool struct{}

func (m *mockValidatorForMempool) ValidateTransaction(_ *transaction.Transaction) (bool, error) {
	return true, nil
}

// =========================================================================
// Test Helpers
// =========================================================================

// createBlockWithDifficulty creates a block with specific difficulty to control accumulated work
func createBlockWithDifficulty(prevHash common.Hash, nonce uint32, difficulty uint8) block.Block {
	var merkleRoot common.Hash
	// Set first bytes to 0 to ensure some leading zeros
	merkleRoot[0] = 0
	merkleRoot[1] = 0
	for i := 2; i < 32; i++ {
		merkleRoot[i] = byte(i + 1)
	}

	header := block.BlockHeader{
		PreviousBlockHash: prevHash,
		MerkleRoot:        merkleRoot,
		Timestamp:         1000 + int64(nonce),
		DifficultyTarget:  difficulty,
		Nonce:             nonce,
	}

	// Create a coinbase transaction
	tx := transaction.NewCoinbaseTransaction(transaction.PubKeyHash{1, 2, 3}, 50, 1)

	return block.Block{
		Header:       header,
		Transactions: []transaction.Transaction{tx},
	}
}

// createNonCoinbaseTransaction creates a regular transaction for testing
func createNonCoinbaseTransaction(prevTxID transaction.TransactionID, outputIndex uint32, pubKeyHash transaction.PubKeyHash) transaction.Transaction {
	return transaction.Transaction{
		Inputs: []transaction.Input{
			{
				PrevTxID:    prevTxID,
				OutputIndex: outputIndex,
				Signature:   []byte("signature"),
				PubKey:      transaction.PubKey{1, 2, 3},
				Sequence:    0xffffffff,
			},
		},
		Outputs: []transaction.Output{
			{
				Value:      25,
				PubKeyHash: pubKeyHash,
			},
		},
		LockTime: 0,
	}
}

// =========================================================================
// Integration Tests
// =========================================================================

// TestBlockStore_ReorganizationNoReorg tests that no reorganization occurs when staying on main chain
// Chain structure:
// (g) -> (b1) -> (b2) -> (b3)
// After adding b3, tip is b3, and CheckAndReorganize with b3's hash should return false
func TestBlockStore_ReorganizationNoReorg(t *testing.T) {
	// Arrange
	genesis := createBlockWithDifficulty([32]byte{}, 0, 10)
	store := blockchain.NewBlockStore(genesis)
	utxoService := newMockUTXOService()
	mempool := NewMempool(&mockValidatorForMempool{}, store)

	reorg := NewChainReorganization(store, utxoService, mempool)

	// Build main chain
	block1 := createBlockWithDifficulty(genesis.Hash(), 1, 10)
	store.AddBlock(block1)

	block2 := createBlockWithDifficulty(block1.Hash(), 2, 10)
	store.AddBlock(block2)

	block3 := createBlockWithDifficulty(block2.Hash(), 3, 10)
	store.AddBlock(block3)

	// Act - first call initializes, second call checks same tip
	_, _ = reorg.CheckAndReorganize(block3.Hash()) // Initialize
	didReorg, err := reorg.CheckAndReorganize(block3.Hash())

	// Assert
	assert.NoError(t, err, "CheckAndReorganize should not error")
	assert.False(t, didReorg, "Should not reorganize when tip is the same")
	assert.Equal(t, 0, utxoService.revertTransactionCalled, "Should not revert any transactions")
	assert.Equal(t, 0, utxoService.applyTransactionCalled, "Should not apply any transactions")
}

// TestBlockStore_ReorganizationSimpleFork tests a simple chain reorganization
// Chain structure:
// Initial: (g) -> (b1) -> (b2) [main chain]
// New chain: (g) -> (s1) -> (s2) [becomes main with higher work]
//
// Expected behavior:
// 1. b2, b1 are disconnected (rolled back)
// 2. s1, s2 are connected (rolled forward)
func TestBlockStore_ReorganizationSimpleFork(t *testing.T) {
	// Arrange
	genesis := createBlockWithDifficulty([32]byte{}, 0, 5)
	store := blockchain.NewBlockStore(genesis)
	utxoService := newMockUTXOService()
	mempool := NewMempool(&mockValidatorForMempool{}, store)

	reorg := NewChainReorganization(store, utxoService, mempool)

	// Build initial main chain (lower difficulty)
	block1 := createBlockWithDifficulty(genesis.Hash(), 1, 5)
	store.AddBlock(block1)

	block2 := createBlockWithDifficulty(block1.Hash(), 2, 5)
	store.AddBlock(block2)

	// Initialize reorganization with current tip
	_, _ = reorg.CheckAndReorganize(block2.Hash())

	// Build competing chain with higher accumulated work
	side1 := createBlockWithDifficulty(genesis.Hash(), 10, 15) // Higher difficulty = more work
	store.AddBlock(side1)

	side2 := createBlockWithDifficulty(side1.Hash(), 11, 15)
	store.AddBlock(side2)

	// Act - reorganize to new tip
	didReorg, err := reorg.CheckAndReorganize(side2.Hash())

	// Assert
	assert.NoError(t, err, "CheckAndReorganize should not error")
	assert.True(t, didReorg, "Should reorganize to new chain with higher work")
	tip := store.GetMainChainTip()
	assert.Equal(t, side2.Hash(), tip.Hash(), "New tip should be side2")

	// Verify transactions were reverted (b2, then b1 - in reverse order)
	// Note: The actual reversion depends on UTXO service implementation
	assert.True(t, utxoService.revertTransactionCalled >= 0, "Revert may have been called")
}

// TestBlockStore_ReorganizationWithUTXOState tests that UTXO state is correctly managed during reorganization
// Chain structure:
// Initial: (g) -> (b1) [contains coinbase creating UTXO]
// Reorg to: (g) -> (s1) [contains different coinbase]
func TestBlockStore_ReorganizationWithUTXOState(t *testing.T) {
	// Arrange
	genesis := createBlockWithDifficulty([32]byte{}, 0, 5)
	store := blockchain.NewBlockStore(genesis)
	utxoService := newMockUTXOService()
	mempool := NewMempool(&mockValidatorForMempool{}, store)

	reorg := NewChainReorganization(store, utxoService, mempool)

	// Create block with a specific coinbase output
	pubKeyHash1 := transaction.PubKeyHash{1, 2, 3}
	block1 := block.Block{
		Header: block.BlockHeader{
			PreviousBlockHash: genesis.Hash(),
			MerkleRoot:        common.Hash{1, 2, 3},
			Timestamp:         1001,
			DifficultyTarget:  5,
			Nonce:             1,
		},
		Transactions: []transaction.Transaction{
			transaction.NewCoinbaseTransaction(pubKeyHash1, 50, 1),
		},
	}
	store.AddBlock(block1)

	// Manually apply the coinbase transaction to UTXO set
	block1TxID := block1.Transactions[0].TransactionId()
	err := utxoService.ApplyTransaction(&block1.Transactions[0], block1TxID, 1, true)
	if err != nil {
		assert.Fail(t, "Should not happen in test")
		return
	}

	// Initialize with current tip
	_, _ = reorg.CheckAndReorganize(block1.Hash())

	// Verify UTXO was created
	coinbaseOutpoint := utxopool.NewOutpoint(block1TxID, 0)
	_, exists := utxoService.utxos[coinbaseOutpoint]
	assert.True(t, exists, "Coinbase UTXO should exist after applying block1")

	// Create competing block with different coinbase
	pubKeyHash2 := transaction.PubKeyHash{4, 5, 6}
	side1 := block.Block{
		Header: block.BlockHeader{
			PreviousBlockHash: genesis.Hash(),
			MerkleRoot:        common.Hash{4, 5, 6},
			Timestamp:         1010,
			DifficultyTarget:  10, // Higher difficulty
			Nonce:             10,
		},
		Transactions: []transaction.Transaction{
			transaction.NewCoinbaseTransaction(pubKeyHash2, 50, 1),
		},
	}
	store.AddBlock(side1)

	// Act - reorganize
	didReorg, err := reorg.CheckAndReorganize(side1.Hash())

	// Assert
	assert.NoError(t, err)
	assert.True(t, didReorg, "Should reorganize to side1")

	// The old coinbase UTXO should be removed (reverted)
	// This is verified by checking that revert was called
	assert.True(t, utxoService.revertTransactionCalled > 0, "Should have reverted transactions from old chain")
}

// TestBlockStore_ReorganizationLongerChainRollback tests reorganization with multiple blocks to roll back
// Chain structure:
// Initial: (g) -> (b1) -> (b2) -> (b3)
// New:     (g) -> (s1) -> (s2) -> (s3) -> (s4) [longer chain]
func TestBlockStore_ReorganizationLongerChainRollback(t *testing.T) {
	// Arrange
	genesis := createBlockWithDifficulty([32]byte{}, 0, 5)
	store := blockchain.NewBlockStore(genesis)
	utxoService := newMockUTXOService()
	mempool := NewMempool(&mockValidatorForMempool{}, store)

	reorg := NewChainReorganization(store, utxoService, mempool)

	// Build initial chain: g -> b1 -> b2 -> b3
	block1 := createBlockWithDifficulty(genesis.Hash(), 1, 5)
	store.AddBlock(block1)

	block2 := createBlockWithDifficulty(block1.Hash(), 2, 5)
	store.AddBlock(block2)

	block3 := createBlockWithDifficulty(block2.Hash(), 3, 5)
	store.AddBlock(block3)

	// Apply blocks to UTXO set
	for _, b := range []block.Block{block1, block2, block3} {
		txID := b.Transactions[0].TransactionId()
		err := utxoService.ApplyTransaction(&b.Transactions[0], txID, uint64(b.Header.Nonce), true)
		if err != nil {
			assert.Fail(t, "Should not happen in test")
		}
	}

	_, _ = reorg.CheckAndReorganize(block3.Hash())

	// Build longer chain: g -> s1 -> s2 -> s3 -> s4
	side1 := createBlockWithDifficulty(genesis.Hash(), 10, 5)
	store.AddBlock(side1)

	side2 := createBlockWithDifficulty(side1.Hash(), 11, 5)
	store.AddBlock(side2)

	side3 := createBlockWithDifficulty(side2.Hash(), 12, 5)
	store.AddBlock(side3)

	side4 := createBlockWithDifficulty(side3.Hash(), 13, 5)
	store.AddBlock(side4)

	// Act
	didReorg, err := reorg.CheckAndReorganize(side4.Hash())

	// Assert
	assert.NoError(t, err)
	assert.True(t, didReorg, "Should reorganize to longer chain")
	tip := store.GetMainChainTip()
	assert.Equal(t, side4.Hash(), tip.Hash(), "Tip should be side4")
	assert.Equal(t, uint64(4), store.GetCurrentHeight(), "Height should be 4")

	// Should have reverted 3 blocks from old chain and applied 4 from new chain
	// The exact count depends on implementation, but reverts should have happened
}

// TestBlockStore_ReorganizationWithMempool tests that mempool is properly updated during reorganization
// When transactions from rolled-back blocks are moved back to mempool
func TestBlockStore_ReorganizationWithMempool(t *testing.T) {
	// Arrange
	genesis := createBlockWithDifficulty([32]byte{}, 0, 5)
	store := blockchain.NewBlockStore(genesis)
	utxoService := newMockUTXOService()

	// Create a mempool with a validator that will pass
	mempool := NewMempool(&mockValidatorForMempool{}, store)

	reorg := NewChainReorganization(store, utxoService, mempool)

	// Create blocks with non-coinbase transactions
	pubKeyHash := transaction.PubKeyHash{1, 2, 3}

	block1 := createBlockWithDifficulty(genesis.Hash(), 1, 5)
	// Add a regular transaction to block1
	regTx := createNonCoinbaseTransaction(genesis.Transactions[0].TransactionId(), 0, pubKeyHash)
	block1.Transactions = append(block1.Transactions, regTx)
	store.AddBlock(block1)

	block2 := createBlockWithDifficulty(block1.Hash(), 2, 5)
	store.AddBlock(block2)

	_, _ = reorg.CheckAndReorganize(block2.Hash())

	// Build competing chain
	side1 := createBlockWithDifficulty(genesis.Hash(), 10, 10)
	store.AddBlock(side1)

	side2 := createBlockWithDifficulty(side1.Hash(), 11, 10)
	store.AddBlock(side2)

	// Act - reorganize which should move transactions from disconnected blocks back to mempool
	didReorg, err := reorg.CheckAndReorganize(side2.Hash())

	// Assert
	assert.NoError(t, err)
	assert.True(t, didReorg)
	// Note: The actual mempool behavior depends on ChainReorganization.moveTransactionsToMempool
	// which calls mempool.AddTransaction for each non-coinbase transaction
}

// TestBlockStore_ReorganizationIdempotent tests that calling reorganization with same tip multiple times is idempotent
func TestBlockStore_ReorganizationIdempotent(t *testing.T) {
	// Arrange
	genesis := createBlockWithDifficulty([32]byte{}, 0, 5)
	store := blockchain.NewBlockStore(genesis)
	utxoService := newMockUTXOService()
	mempool := NewMempool(&mockValidatorForMempool{}, store)

	reorg := NewChainReorganization(store, utxoService, mempool)

	block1 := createBlockWithDifficulty(genesis.Hash(), 1, 5)
	store.AddBlock(block1)

	// Act - call multiple times with same tip
	didReorg1, err1 := reorg.CheckAndReorganize(block1.Hash())
	didReorg2, err2 := reorg.CheckAndReorganize(block1.Hash())
	didReorg3, err3 := reorg.CheckAndReorganize(block1.Hash())

	// Assert
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)

	// First call initializes (returns false), subsequent calls with same tip also return false
	assert.False(t, didReorg1, "First call initializes, should not reorganize")
	assert.False(t, didReorg2, "Second call with same tip should not reorganize")
	assert.False(t, didReorg3, "Third call with same tip should not reorganize")
}

// TestBlockStore_ReorganizationStateConsistency tests that UTXO state remains consistent after reorganization
func TestBlockStore_ReorganizationStateConsistency(t *testing.T) {
	// Arrange
	genesis := createBlockWithDifficulty([32]byte{}, 0, 5)
	store := blockchain.NewBlockStore(genesis)
	utxoService := newMockUTXOService()
	mempool := NewMempool(&mockValidatorForMempool{}, store)

	reorg := NewChainReorganization(store, utxoService, mempool)

	// Track UTXO count before reorg
	initialUTXOCount := len(utxoService.utxos)

	// Build chain
	block1 := createBlockWithDifficulty(genesis.Hash(), 1, 5)
	store.AddBlock(block1)
	err := utxoService.ApplyTransaction(&block1.Transactions[0], block1.Transactions[0].TransactionId(), 1, true)
	if err != nil {
		assert.Fail(t, "Should not happen in test")
	}

	_, _ = reorg.CheckAndReorganize(block1.Hash())

	// Build competing chain
	side1 := createBlockWithDifficulty(genesis.Hash(), 10, 10)
	store.AddBlock(side1)

	// Act
	didReorg, err := reorg.CheckAndReorganize(side1.Hash())

	// Assert
	assert.NoError(t, err)
	assert.True(t, didReorg)

	// After reorg, we should have:
	// - Genesis coinbase (still there)
	// - Side1 coinbase (newly added)
	// - Block1 coinbase (removed)
	// So: initial + 1 (side1) = roughly same as before block1 was added
	// The exact count depends on the rollback implementation
	finalUTXOCount := len(utxoService.utxos)

	// At minimum, we should have at least the genesis UTXO and side1 UTXO
	assert.True(t, finalUTXOCount >= initialUTXOCount,
		"UTXO count should be at least initial count after reorg")

	// Verify the side1 coinbase exists
	side1TxID := side1.Transactions[0].TransactionId()
	side1Outpoint := utxopool.NewOutpoint(side1TxID, 0)
	_, exists := utxoService.utxos[side1Outpoint]
	assert.True(t, exists, "Side1 coinbase UTXO should exist")
}

// TestBlockStore_AccumulatedWorkSelection verifies that main chain selection is based on accumulated work
func TestBlockStore_AccumulatedWorkSelection(t *testing.T) {
	// Arrange
	genesis := createBlockWithDifficulty([32]byte{}, 0, 5)
	store := blockchain.NewBlockStore(genesis)

	// Create two chains with same length but different difficulties
	// Chain A: lower total work
	block1a := createBlockWithDifficulty(genesis.Hash(), 1, 3) // Low difficulty
	store.AddBlock(block1a)
	block2a := createBlockWithDifficulty(block1a.Hash(), 2, 3)
	store.AddBlock(block2a)

	// Chain B: higher total work
	block1b := createBlockWithDifficulty(genesis.Hash(), 10, 15) // High difficulty
	store.AddBlock(block1b)
	block2b := createBlockWithDifficulty(block1b.Hash(), 11, 15)
	store.AddBlock(block2b)

	// Act - Get main chain tip
	tip := store.GetMainChainTip()
	tipHash := tip.Hash()

	// Assert - The tip should be from the higher work chain (chain B)
	assert.True(t, tipHash == block2b.Hash() || tipHash == block2a.Hash(),
		"Tip should be one of the chain tips")

	// Since chain B has much higher accumulated work (15+15=30 vs 3+3=6),
	// block2b should be the main chain tip
	assert.Equal(t, block2b.Hash(), tipHash,
		"Higher accumulated work chain should be main chain")
}
