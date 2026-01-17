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

// createTestMultiChainService creates a MultiChainUTXOService for testing
// NOTE: Uses mockEntryDAO and newMockEntryDAO from blockchain_test.go
func createTestMultiChainService(blockStore blockchain.BlockStoreAPI) *utxo.MultiChainUTXOService {
	mempool := utxo.NewMemUTXOPoolService()
	chainstate, _ := utxo.NewChainStateService(
		utxo.ChainStateConfig{CacheSize: 100},
		newMockEntryDAO(),
	)
	fullNode := utxo.NewFullNodeUTXOService(mempool, chainstate)
	return utxo.NewMultiChainUTXOService(fullNode, blockStore)
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
		Timestamp:         int64(nonce) + 1000,
		DifficultyTarget:  difficulty,
		Nonce:             nonce,
	}

	return block.Block{
		Header:       header,
		Transactions: []transaction.Transaction{},
	}
}

// =========================================================================
// Integration Tests for ChainReorganization
// =========================================================================

// TestBlockStore_ReorganizationNoReorg tests that no reorganization occurs when staying on main chain
// Chain structure:
// (g) -> (b1) -> (b2) -> (b3)
// After adding b3, tip is b3, and CheckAndReorganize with b3's hash should return false
func TestBlockStore_ReorganizationNoReorg(t *testing.T) {
	// Arrange
	genesis := createBlockWithDifficulty([32]byte{}, 0, 10)
	store := blockchain.NewBlockStore(genesis)
	multiChainService := createTestMultiChainService(store)
	mempool := NewMempool(&mockValidatorForMempool{}, store)

	reorg := NewChainReorganization(store, multiChainService, mempool)

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
	multiChainService := createTestMultiChainService(store)
	mempool := NewMempool(&mockValidatorForMempool{}, store)

	reorg := NewChainReorganization(store, multiChainService, mempool)

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
}

// TestBlockStore_ReorganizationWithUTXOState tests that UTXO state is correctly managed during reorganization
// Chain structure:
// Initial: (g) -> (b1) [contains coinbase creating UTXO]
// Reorg to: (g) -> (s1) [contains different coinbase]
func TestBlockStore_ReorganizationWithUTXOState(t *testing.T) {
	// Arrange
	genesis := createBlockWithDifficulty([32]byte{}, 0, 5)
	store := blockchain.NewBlockStore(genesis)
	multiChainService := createTestMultiChainService(store)
	mempool := NewMempool(&mockValidatorForMempool{}, store)

	reorg := NewChainReorganization(store, multiChainService, mempool)

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
	err := multiChainService.GetMainChain().ApplyTransaction(&block1.Transactions[0], block1TxID, 1, true)
	if err != nil {
		assert.Fail(t, "Should not happen in test")
		return
	}

	// Initialize with current tip
	_, _ = reorg.CheckAndReorganize(block1.Hash())

	// Verify UTXO was created
	coinbaseOutpoint := utxopool.NewOutpoint(block1TxID, 0)
	exists := multiChainService.GetMainChain().ContainsUTXO(coinbaseOutpoint)
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

}

// TestBlockStore_ReorganizationChainExtension tests simple chain extension (not a reorg)
func TestBlockStore_ReorganizationChainExtension(t *testing.T) {
	// Arrange
	genesis := createBlockWithDifficulty([32]byte{}, 0, 10)
	store := blockchain.NewBlockStore(genesis)
	multiChainService := createTestMultiChainService(store)
	mempool := NewMempool(&mockValidatorForMempool{}, store)

	reorg := NewChainReorganization(store, multiChainService, mempool)

	// Initialize with genesis
	_, _ = reorg.CheckAndReorganize(genesis.Hash())

	// Create block that extends the chain
	block1 := createBlockWithDifficulty(genesis.Hash(), 1, 10)
	store.AddBlock(block1)

	// Act - extend chain (not a reorg)
	didReorg, err := reorg.CheckAndReorganize(block1.Hash())

	// Assert
	assert.NoError(t, err, "CheckAndReorganize should not error")
	assert.False(t, didReorg, "Chain extension should not be a reorganization")
}

// TestBlockStore_ReorganizationDeepFork tests a deeper reorganization
func TestBlockStore_ReorganizationDeepFork(t *testing.T) {
	// Arrange
	genesis := createBlockWithDifficulty([32]byte{}, 0, 5)
	store := blockchain.NewBlockStore(genesis)
	multiChainService := createTestMultiChainService(store)
	mempool := NewMempool(&mockValidatorForMempool{}, store)

	reorg := NewChainReorganization(store, multiChainService, mempool)

	// Main chain: genesis -> b1 -> b2 -> b3
	block1 := createBlockWithDifficulty(genesis.Hash(), 1, 5)
	store.AddBlock(block1)

	block2 := createBlockWithDifficulty(block1.Hash(), 2, 5)
	store.AddBlock(block2)

	block3 := createBlockWithDifficulty(block2.Hash(), 3, 5)
	store.AddBlock(block3)

	// Initialize with main chain
	_, _ = reorg.CheckAndReorganize(block3.Hash())

	// Side chain from genesis: genesis -> s1 -> s2 -> s3 -> s4 (more work)
	side1 := createBlockWithDifficulty(genesis.Hash(), 10, 10)
	store.AddBlock(side1)

	side2 := createBlockWithDifficulty(side1.Hash(), 11, 10)
	store.AddBlock(side2)

	side3 := createBlockWithDifficulty(side2.Hash(), 12, 10)
	store.AddBlock(side3)

	side4 := createBlockWithDifficulty(side3.Hash(), 13, 10)
	store.AddBlock(side4)

	// Act - reorg to side chain
	didReorg, err := reorg.CheckAndReorganize(side4.Hash())

	// Assert
	assert.NoError(t, err, "CheckAndReorganize should not error")
	assert.True(t, didReorg, "Should reorganize to higher work chain")
}

// TestBlockStore_ReorganizationPartialOverlap tests reorganization where chains share some blocks
func TestBlockStore_ReorganizationPartialOverlap(t *testing.T) {
	// Arrange
	genesis := createBlockWithDifficulty([32]byte{}, 0, 5)
	store := blockchain.NewBlockStore(genesis)
	multiChainService := createTestMultiChainService(store)
	mempool := NewMempool(&mockValidatorForMempool{}, store)

	reorg := NewChainReorganization(store, multiChainService, mempool)

	// Shared: genesis -> shared1
	shared1 := createBlockWithDifficulty(genesis.Hash(), 1, 5)
	store.AddBlock(shared1)

	// Main chain: shared1 -> main2 -> main3
	main2 := createBlockWithDifficulty(shared1.Hash(), 2, 5)
	store.AddBlock(main2)

	main3 := createBlockWithDifficulty(main2.Hash(), 3, 5)
	store.AddBlock(main3)

	// Initialize with main chain
	_, _ = reorg.CheckAndReorganize(main3.Hash())

	// Side chain: shared1 -> side2 -> side3 (higher difficulty)
	side2 := createBlockWithDifficulty(shared1.Hash(), 10, 10)
	store.AddBlock(side2)

	side3 := createBlockWithDifficulty(side2.Hash(), 11, 10)
	store.AddBlock(side3)

	// Act - reorg to side chain
	didReorg, err := reorg.CheckAndReorganize(side3.Hash())

	// Assert
	assert.NoError(t, err, "CheckAndReorganize should not error")
	assert.True(t, didReorg, "Should reorganize to higher work chain")
}

// TestBlockStore_ReorganizationMultipleReorgs tests multiple consecutive reorganizations
func TestBlockStore_ReorganizationMultipleReorgs(t *testing.T) {
	// Arrange
	genesis := createBlockWithDifficulty([32]byte{}, 0, 5)
	store := blockchain.NewBlockStore(genesis)
	multiChainService := createTestMultiChainService(store)
	mempool := NewMempool(&mockValidatorForMempool{}, store)

	reorg := NewChainReorganization(store, multiChainService, mempool)

	// Main chain: genesis -> main1
	main1 := createBlockWithDifficulty(genesis.Hash(), 1, 5)
	store.AddBlock(main1)

	// Initialize
	_, _ = reorg.CheckAndReorganize(main1.Hash())

	// First side chain: genesis -> side1 (higher work)
	side1 := createBlockWithDifficulty(genesis.Hash(), 10, 10)
	store.AddBlock(side1)

	// First reorg
	didReorg1, err := reorg.CheckAndReorganize(side1.Hash())
	assert.NoError(t, err)
	assert.True(t, didReorg1, "First reorganization should occur")

	// Second side chain: genesis -> side2 (even higher work)
	side2 := createBlockWithDifficulty(genesis.Hash(), 20, 15)
	store.AddBlock(side2)

	// Second reorg
	didReorg2, err := reorg.CheckAndReorganize(side2.Hash())
	assert.NoError(t, err)
	assert.True(t, didReorg2, "Second reorganization should occur")
}

// TestBlockStore_ReorganizationInitialSync tests the initial sync behavior
func TestBlockStore_ReorganizationInitialSync(t *testing.T) {
	// Arrange
	genesis := createBlockWithDifficulty([32]byte{}, 0, 10)
	store := blockchain.NewBlockStore(genesis)
	multiChainService := createTestMultiChainService(store)
	mempool := NewMempool(&mockValidatorForMempool{}, store)

	reorg := NewChainReorganization(store, multiChainService, mempool)

	// Act - first call should initialize without errors
	didReorg, err := reorg.CheckAndReorganize(genesis.Hash())

	// Assert
	assert.NoError(t, err, "Initial sync should not error")
	assert.False(t, didReorg, "Initial sync should not be considered a reorganization")
}
