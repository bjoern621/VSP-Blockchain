package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/blockchain"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"testing"

	"github.com/stretchr/testify/assert"
)

// =========================================================================
// Mock Implementations for Integration Testing
// =========================================================================

// Outpoint uniquely identifies a UTXO by transaction ID and output index (mirrors utxo.Outpoint)
type testOutpoint struct {
	TxID        transaction.TransactionID
	OutputIndex uint32
}

// mockUTXOService is a mock for UtxoStoreAPI that tracks state changes
type mockUTXOService struct {
	// blockHashToPool maps block hash to its UTXO pool
	blockHashToPool map[common.Hash]map[testOutpoint]transaction.Output
	blockStore      blockchain.BlockStoreAPI

	// Tracking for test assertions
	addNewBlockCalled           int
	initializeGenesisPoolCalled int
}

func newMockUTXOService(blockStore blockchain.BlockStoreAPI) *mockUTXOService {
	return &mockUTXOService{
		blockHashToPool: make(map[common.Hash]map[testOutpoint]transaction.Output),
		blockStore:      blockStore,
	}
}

// InitializeGenesisPool creates the UTXO pool for the genesis block.
func (m *mockUTXOService) InitializeGenesisPool(genesisBlock block.Block) error {
	m.initializeGenesisPoolCalled++

	genesisHash := genesisBlock.Hash()
	if _, exists := m.blockHashToPool[genesisHash]; exists {
		return nil
	}

	// Create genesis pool from empty previous pool
	genesisPool := make(map[testOutpoint]transaction.Output)

	// Add all outputs from genesis transactions
	for _, tx := range genesisBlock.Transactions {
		txID := tx.TransactionId()
		for i, output := range tx.Outputs {
			outpoint := testOutpoint{TxID: txID, OutputIndex: uint32(i)}
			genesisPool[outpoint] = output
		}
	}

	m.blockHashToPool[genesisHash] = genesisPool
	return nil
}

// AddNewBlock creates a new UTXO pool for the block.
func (m *mockUTXOService) AddNewBlock(newBlock block.Block) error {
	m.addNewBlockCalled++

	newBlockHash := newBlock.Hash()

	// Skip if already exists
	if _, exists := m.blockHashToPool[newBlockHash]; exists {
		return nil
	}

	// Get previous pool
	prevBlockHash := newBlock.Header.PreviousBlockHash
	prevPool, exists := m.blockHashToPool[prevBlockHash]
	if !exists {
		prevPool = make(map[testOutpoint]transaction.Output)
	}

	// Create new pool by copying previous
	newPool := make(map[testOutpoint]transaction.Output)
	for outpoint, output := range prevPool {
		newPool[outpoint] = output
	}

	// Remove spent UTXOs
	for _, tx := range newBlock.Transactions {
		if tx.IsCoinbase() {
			continue
		}
		for _, input := range tx.Inputs {
			outpoint := testOutpoint{TxID: input.PrevTxID, OutputIndex: input.OutputIndex}
			delete(newPool, outpoint)
		}
	}

	// Add new outputs
	for _, tx := range newBlock.Transactions {
		txID := tx.TransactionId()
		for i, output := range tx.Outputs {
			outpoint := testOutpoint{TxID: txID, OutputIndex: uint32(i)}
			newPool[outpoint] = output
		}
	}

	m.blockHashToPool[newBlockHash] = newPool
	return nil
}

// GetUtxoFromBlock retrieves a specific UTXO from a given block's UTXO pool.
func (m *mockUTXOService) GetUtxoFromBlock(prevTxID transaction.TransactionID, outputIndex uint32, blockHash common.Hash) (transaction.Output, error) {
	pool, exists := m.blockHashToPool[blockHash]
	if !exists {
		return transaction.Output{}, nil
	}
	outpoint := testOutpoint{TxID: prevTxID, OutputIndex: outputIndex}
	output, exists := pool[outpoint]
	if !exists {
		return transaction.Output{}, nil
	}
	return output, nil
}

// ValidateBlock checks if all inputs reference valid UTXOs.
func (m *mockUTXOService) ValidateBlock(blockToValidate block.Block) bool {
	prevBlockHash := blockToValidate.Header.PreviousBlockHash
	prevPool, exists := m.blockHashToPool[prevBlockHash]
	if !exists {
		return false
	}

	for _, tx := range blockToValidate.Transactions {
		if tx.IsCoinbase() {
			continue
		}
		for _, input := range tx.Inputs {
			outpoint := testOutpoint{TxID: input.PrevTxID, OutputIndex: input.OutputIndex}
			if _, exists := prevPool[outpoint]; !exists {
				return false
			}
		}
	}
	return true
}

// ValidateTransactionFromBlock checks if a transaction is valid against the UTXO set at a specific block.
func (m *mockUTXOService) ValidateTransactionFromBlock(tx transaction.Transaction, blockHash common.Hash) bool {
	pool, exists := m.blockHashToPool[blockHash]
	if !exists {
		return false
	}

	for _, input := range tx.Inputs {
		outpoint := testOutpoint{TxID: input.PrevTxID, OutputIndex: input.OutputIndex}
		if _, exists := pool[outpoint]; !exists {
			return false
		}
	}
	return true
}

// GetUtxosByPubKeyHashFromBlock retrieves all UTXOs associated with a public key hash.
func (m *mockUTXOService) GetUtxosByPubKeyHashFromBlock(pubKeyHash transaction.PubKeyHash, blockHash common.Hash) ([]transaction.UTXO, error) {
	pool, exists := m.blockHashToPool[blockHash]
	if !exists {
		return nil, nil
	}

	utxos := make([]transaction.UTXO, 0)
	for outpoint, output := range pool {
		if output.PubKeyHash == pubKeyHash {
			utxo := transaction.UTXO{
				TxID:        outpoint.TxID,
				OutputIndex: outpoint.OutputIndex,
				Output:      output,
			}
			utxos = append(utxos, utxo)
		}
	}
	return utxos, nil
}

// mockValidatorForMempool is a mock that always validates successfully
type mockValidatorForMempool struct{}

func (m *mockValidatorForMempool) ValidateTransaction(_ transaction.Transaction, _ common.Hash) (bool, error) {
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
			},
		},
		Outputs: []transaction.Output{
			{
				Value:      25,
				PubKeyHash: pubKeyHash,
			},
		},
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
	utxoService := newMockUTXOService(store)
	_ = utxoService.InitializeGenesisPool(genesis)
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
	utxoService := newMockUTXOService(store)
	_ = utxoService.InitializeGenesisPool(genesis)
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

	// Verify UTXO pools were created for the new chain
	assert.True(t, utxoService.addNewBlockCalled >= 2, "Should have added blocks to UTXO store")
}

// TestBlockStore_ReorganizationWithUTXOState tests that UTXO state is correctly managed during reorganization
// Chain structure:
// Initial: (g) -> (b1) [contains coinbase creating UTXO]
// Reorg to: (g) -> (s1) [contains different coinbase]
func TestBlockStore_ReorganizationWithUTXOState(t *testing.T) {
	// Arrange
	genesis := createBlockWithDifficulty([32]byte{}, 0, 5)
	store := blockchain.NewBlockStore(genesis)
	utxoService := newMockUTXOService(store)
	_ = utxoService.InitializeGenesisPool(genesis)
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

	// Initialize with current tip (this will add block1 to UTXO store)
	_, _ = reorg.CheckAndReorganize(block1.Hash())

	// Verify UTXO was created for block1
	block1TxID := block1.Transactions[0].TransactionId()
	block1Pool, exists := utxoService.blockHashToPool[block1.Hash()]
	assert.True(t, exists, "Block1 UTXO pool should exist")
	coinbaseOutpoint := testOutpoint{TxID: block1TxID, OutputIndex: 0}
	_, utxoExists := block1Pool[coinbaseOutpoint]
	assert.True(t, utxoExists, "Coinbase UTXO should exist in block1 pool")

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

	// The side1 coinbase UTXO should exist in its pool
	side1TxID := side1.Transactions[0].TransactionId()
	side1Pool, exists := utxoService.blockHashToPool[side1.Hash()]
	assert.True(t, exists, "Side1 UTXO pool should exist")
	side1Outpoint := testOutpoint{TxID: side1TxID, OutputIndex: 0}
	_, side1UtxoExists := side1Pool[side1Outpoint]
	assert.True(t, side1UtxoExists, "Side1 coinbase UTXO should exist in its pool")
}

// TestBlockStore_ReorganizationLongerChainRollback tests reorganization with multiple blocks to roll back
// Chain structure:
// Initial: (g) -> (b1) -> (b2) -> (b3)
// New:     (g) -> (s1) -> (s2) -> (s3) -> (s4) [longer chain]
func TestBlockStore_ReorganizationLongerChainRollback(t *testing.T) {
	// Arrange
	genesis := createBlockWithDifficulty([32]byte{}, 0, 5)
	store := blockchain.NewBlockStore(genesis)
	utxoService := newMockUTXOService(store)
	_ = utxoService.InitializeGenesisPool(genesis)
	mempool := NewMempool(&mockValidatorForMempool{}, store)

	reorg := NewChainReorganization(store, utxoService, mempool)

	// Build initial chain: g -> b1 -> b2 -> b3
	block1 := createBlockWithDifficulty(genesis.Hash(), 1, 5)
	store.AddBlock(block1)

	block2 := createBlockWithDifficulty(block1.Hash(), 2, 5)
	store.AddBlock(block2)

	block3 := createBlockWithDifficulty(block2.Hash(), 3, 5)
	store.AddBlock(block3)

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

	// Verify UTXO pools were created for all new chain blocks
	_, exists := utxoService.blockHashToPool[side4.Hash()]
	assert.True(t, exists, "Side4 UTXO pool should exist")
}

// TestBlockStore_ReorganizationWithMempool tests that mempool is properly updated during reorganization
// When transactions from rolled-back blocks are moved back to mempool
func TestBlockStore_ReorganizationWithMempool(t *testing.T) {
	// Arrange
	genesis := createBlockWithDifficulty([32]byte{}, 0, 5)
	store := blockchain.NewBlockStore(genesis)
	utxoService := newMockUTXOService(store)
	_ = utxoService.InitializeGenesisPool(genesis)

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
	utxoService := newMockUTXOService(store)
	_ = utxoService.InitializeGenesisPool(genesis)
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
	utxoService := newMockUTXOService(store)
	_ = utxoService.InitializeGenesisPool(genesis)
	mempool := NewMempool(&mockValidatorForMempool{}, store)

	reorg := NewChainReorganization(store, utxoService, mempool)

	// Build chain
	block1 := createBlockWithDifficulty(genesis.Hash(), 1, 5)
	store.AddBlock(block1)

	_, _ = reorg.CheckAndReorganize(block1.Hash())

	// Verify block1 UTXO pool was created
	block1Pool, exists := utxoService.blockHashToPool[block1.Hash()]
	assert.True(t, exists, "Block1 UTXO pool should exist")
	block1TxID := block1.Transactions[0].TransactionId()
	block1Outpoint := testOutpoint{TxID: block1TxID, OutputIndex: 0}
	_, block1UtxoExists := block1Pool[block1Outpoint]
	assert.True(t, block1UtxoExists, "Block1 coinbase UTXO should exist")

	// Build competing chain
	side1 := createBlockWithDifficulty(genesis.Hash(), 10, 10)
	store.AddBlock(side1)

	// Act
	didReorg, err := reorg.CheckAndReorganize(side1.Hash())

	// Assert
	assert.NoError(t, err)
	assert.True(t, didReorg)

	// Verify the side1 coinbase exists in its pool
	side1TxID := side1.Transactions[0].TransactionId()
	side1Pool, exists := utxoService.blockHashToPool[side1.Hash()]
	assert.True(t, exists, "Side1 UTXO pool should exist")
	side1Outpoint := testOutpoint{TxID: side1TxID, OutputIndex: 0}
	_, side1UtxoExists := side1Pool[side1Outpoint]
	assert.True(t, side1UtxoExists, "Side1 coinbase UTXO should exist")

	// Both pools should still exist (immutable snapshots)
	_, block1PoolStillExists := utxoService.blockHashToPool[block1.Hash()]
	assert.True(t, block1PoolStillExists, "Block1 UTXO pool should still exist (immutable)")
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
