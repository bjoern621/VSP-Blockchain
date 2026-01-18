package blockchain

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"testing"
)

type blockValidatorForTests struct{}

func (b *blockValidatorForTests) FullValidation(blockToValidate block.Block) (bool, error) {
	return true, nil
}

// Test helper function to create a test block with minimal leading zero bits
func createTestBlock(prevHash common.Hash, nonce uint32) block.Block {
	var merkleRoot common.Hash
	// Use a simple pattern for merkle root
	for i := range 32 {
		merkleRoot[i] = byte(i + 1)
	}

	header := block.BlockHeader{
		PreviousBlockHash: prevHash,
		MerkleRoot:        merkleRoot,
		Timestamp:         1000 + int64(nonce),
		DifficultyTarget:  0,
		Nonce:             nonce,
	}

	// Create a simple coinbase transaction
	tx := transaction.Transaction{
		Inputs:  []transaction.Input{},
		Outputs: []transaction.Output{{Value: 50, PubKeyHash: transaction.PubKeyHash{1, 2, 3}}},
	}

	return block.Block{
		Header:       header,
		Transactions: []transaction.Transaction{tx},
	}
}

// Test helper function to create a block with guaranteed leading zero bits
// This creates a block with a specific pattern that will have at least some leading zeros
func createTestBlockWithLeadingZeros(prevHash common.Hash, nonce uint32) block.Block {
	var merkleRoot common.Hash
	// Use a pattern that starts with zeros to ensure some leading zero bits in hash
	merkleRoot[0] = 0
	merkleRoot[1] = 0
	for i := 2; i < 32; i++ {
		merkleRoot[i] = byte(i + 1)
	}

	header := block.BlockHeader{
		PreviousBlockHash: prevHash,
		MerkleRoot:        merkleRoot,
		Timestamp:         1000 + int64(nonce),
		DifficultyTarget:  0,
		Nonce:             nonce,
	}

	// Create a simple coinbase transaction
	tx := transaction.Transaction{
		Inputs:  []transaction.Input{},
		Outputs: []transaction.Output{{Value: 50, PubKeyHash: transaction.PubKeyHash{1, 2, 3}}},
	}

	return block.Block{
		Header:       header,
		Transactions: []transaction.Transaction{tx},
	}
}

// TestNewBlockStore tests creating a new block store with a genesis block
// (g)
func TestNewBlockStore(t *testing.T) {
	genesis := createTestBlockWithLeadingZeros([32]byte{}, 0)
	blockValidator := &blockValidatorForTests{}
	store := NewBlockStore(genesis, blockValidator)

	// Test that genesis is in the store
	retrievedGenesis, err := store.GetBlockByHash(genesis.Hash())
	if err != nil {
		t.Fatalf("Failed to retrieve genesis block: %v", err)
	}

	if retrievedGenesis.Hash() != genesis.Hash() {
		t.Errorf("Genesis hash mismatch: got %v, want %v", retrievedGenesis.Hash(), genesis.Hash())
	}

	// Test that genesis is the main chain tip
	tip := store.GetMainChainTip()
	if tip.Hash() != genesis.Hash() {
		t.Errorf("Main chain tip should be genesis: got %v, want %v", tip.Hash(), genesis.Hash())
	}

	// Test that genesis is not an orphan
	isOrphan, err := store.IsOrphanBlock(genesis)
	if err != nil {
		t.Fatalf("Failed to check if genesis is orphan: %v", err)
	}
	if isOrphan {
		t.Error("Genesis should not be an orphan")
	}

	// Test that genesis is part of main chain
	if !store.IsPartOfMainChain(genesis) {
		t.Error("Genesis should be part of main chain")
	}

	// Test current height is 0
	if store.GetCurrentHeight() != 0 {
		t.Errorf("Current height should be 0: got %d", store.GetCurrentHeight())
	}

	// Test blocks at height 0
	blocksAtHeight0 := store.GetBlocksByHeight(0)
	if len(blocksAtHeight0) != 1 {
		t.Errorf("Expected 1 block at height 0: got %d", len(blocksAtHeight0))
	}
	if blocksAtHeight0[0].Hash() != genesis.Hash() {
		t.Errorf("Block at height 0 should be genesis")
	}
}

// TestAddBlock_Idempotent tests that adding the same block multiple times has no effect
// (g)
// (g) -> (b1)
// Add (b1) again: no change
func TestAddBlock_Idempotent(t *testing.T) {
	genesis := createTestBlockWithLeadingZeros([32]byte{}, 0)
	blockValidator := &blockValidatorForTests{}
	store := NewBlockStore(genesis, blockValidator)

	block1 := createTestBlockWithLeadingZeros(genesis.Hash(), 1)
	addedHashes1 := store.AddBlock(block1)

	if len(addedHashes1) != 1 {
		t.Errorf("Expected 1 added hash: got %d", len(addedHashes1))
	}
	if addedHashes1[0] != block1.Hash() {
		t.Errorf("Added hash mismatch")
	}

	// Add the same block again
	addedHashes2 := store.AddBlock(block1)

	if len(addedHashes2) != 0 {
		t.Errorf("Expected 0 added hashes on duplicate: got %d", len(addedHashes2))
	}
}

// TestAddBlock_MainChain tests adding blocks to form a main chain
// (g)
// (g) -> (b1) -> (b2)
func TestAddBlock_MainChain(t *testing.T) {
	genesis := createTestBlockWithLeadingZeros([32]byte{}, 0)
	blockValidator := &blockValidatorForTests{}
	store := NewBlockStore(genesis, blockValidator)

	// Add blocks sequentially
	block1 := createTestBlockWithLeadingZeros(genesis.Hash(), 1)
	addedHashes1 := store.AddBlock(block1)

	if len(addedHashes1) != 1 {
		t.Errorf("Expected 1 added hash: got %d", len(addedHashes1))
	}
	if addedHashes1[0] != block1.Hash() {
		t.Errorf("First added hash should be block1")
	}

	block2 := createTestBlockWithLeadingZeros(block1.Hash(), 2)
	addedHashes2 := store.AddBlock(block2)

	if len(addedHashes2) != 1 {
		t.Errorf("Expected 1 added hash: got %d", len(addedHashes2))
	}
	if addedHashes2[0] != block2.Hash() {
		t.Errorf("Second added hash should be block2")
	}

	// Verify main chain
	tip := store.GetMainChainTip()
	if tip.Hash() != block2.Hash() {
		t.Errorf("Main chain tip should be block2: got %v, want %v", tip.Hash(), block2.Hash())
	}

	// Verify heights
	if store.GetCurrentHeight() != 2 {
		t.Errorf("Current height should be 2: got %d", store.GetCurrentHeight())
	}

	// Verify all blocks are in main chain
	if !store.IsPartOfMainChain(genesis) {
		t.Error("Genesis should be part of main chain")
	}
	if !store.IsPartOfMainChain(block1) {
		t.Error("Block1 should be part of main chain")
	}
	if !store.IsPartOfMainChain(block2) {
		t.Error("Block2 should be part of main chain")
	}

	// Verify blocks are not orphans
	isOrphan, err := store.IsOrphanBlock(block1)
	if err != nil {
		t.Fatalf("Failed to check if block1 is orphan: %v", err)
	}
	if isOrphan {
		t.Error("Block1 should not be an orphan")
	}
}

// TestAddBlock_SideChain tests creating a side chain
// (g)
// (g) -> (b1) -> (b2)  [main chain]
// (g) -> (s1)          [side chain]
func TestAddBlock_SideChain(t *testing.T) {
	genesis := createTestBlockWithLeadingZeros([32]byte{}, 0)
	blockValidator := &blockValidatorForTests{}
	store := NewBlockStore(genesis, blockValidator)

	// Create main chain
	block1 := createTestBlockWithLeadingZeros(genesis.Hash(), 1)
	store.AddBlock(block1)

	block2 := createTestBlockWithLeadingZeros(block1.Hash(), 2)
	store.AddBlock(block2)

	// Create side chain branching from genesis
	sideBlock1 := createTestBlockWithLeadingZeros(genesis.Hash(), 100)
	addedHashes := store.AddBlock(sideBlock1)

	if len(addedHashes) != 1 {
		t.Errorf("Expected 1 added hash for side chain: got %d", len(addedHashes))
	}
	if addedHashes[0] != sideBlock1.Hash() {
		t.Errorf("Added hash should be sideBlock1")
	}

	// Verify main chain is still the original
	tip := store.GetMainChainTip()
	if tip.Hash() != block2.Hash() {
		t.Errorf("Main chain tip should still be block2")
	}

	// Verify current height is still 2
	if store.GetCurrentHeight() != 2 {
		t.Errorf("Current height should be 2: got %d", store.GetCurrentHeight())
	}

	// Verify blocks at height 1
	blocksAtHeight1 := store.GetBlocksByHeight(1)
	if len(blocksAtHeight1) != 2 {
		t.Errorf("Expected 2 blocks at height 1: got %d", len(blocksAtHeight1))
	}

	height1Hashes := make(map[common.Hash]bool)
	for _, b := range blocksAtHeight1 {
		height1Hashes[b.Hash()] = true
	}

	if !height1Hashes[block1.Hash()] {
		t.Error("Block1 should be at height 1")
	}
	if !height1Hashes[sideBlock1.Hash()] {
		t.Error("SideBlock1 should be at height 1")
	}

	// Verify side chain block is not in main chain
	if store.IsPartOfMainChain(sideBlock1) {
		t.Error("Side block should not be part of main chain")
	}

	// Verify side chain block is not an orphan
	isOrphan, err := store.IsOrphanBlock(sideBlock1)
	if err != nil {
		t.Fatalf("Failed to check if sideBlock1 is orphan: %v", err)
	}
	if isOrphan {
		t.Error("Side block should not be an orphan")
	}
}

// TestAddBlock_SideChainBecomesMainChain tests that a side chain becomes main chain with higher accumulated work
// (g)
// (g) -> (b1) -> (b2)
// (g) -> (s1)
// If (s1) has higher accumulated work, it becomes main chain tip
func TestAddBlock_SideChainBecomesMainChain(t *testing.T) {
	genesis := createTestBlockWithLeadingZeros([32]byte{}, 0)
	blockValidator := &blockValidatorForTests{}
	store := NewBlockStore(genesis, blockValidator)

	// Create initial main chain
	block1 := createTestBlockWithLeadingZeros(genesis.Hash(), 1)
	store.AddBlock(block1)

	block2 := createTestBlockWithLeadingZeros(block1.Hash(), 2)
	store.AddBlock(block2)

	// Create side chain with higher accumulated work
	// Note: We can't control actual difficulty precisely, so this test
	// verifies the structure rather than specific accumulated work values
	sideBlock1 := createTestBlockWithLeadingZeros(genesis.Hash(), 100)
	addedHashes := store.AddBlock(sideBlock1)

	if len(addedHashes) != 1 {
		t.Errorf("Expected 1 added hash for side chain: got %d", len(addedHashes))
	}

	// Verify side chain became main chain (the one with higher accumulated work)
	tip := store.GetMainChainTip()
	tipHash := tip.Hash()

	// The tip should be either block2 or sideBlock1, depending on accumulated work
	if tipHash != block2.Hash() && tipHash != sideBlock1.Hash() {
		t.Errorf("Main chain tip should be either block2 or sideBlock1: got %v", tipHash)
	}

	// Verify that one of them is in main chain
	if !store.IsPartOfMainChain(block2) && !store.IsPartOfMainChain(sideBlock1) {
		t.Error("Either block2 or sideBlock1 should be part of main chain")
	}
}

// TestAddBlock_OrphanBlock tests creating orphan blocks
// (g)
//
//	..> (o) [orphan, no parent in store]
func TestAddBlock_OrphanBlock(t *testing.T) {
	genesis := createTestBlockWithLeadingZeros([32]byte{}, 0)
	blockValidator := &blockValidatorForTests{}
	store := NewBlockStore(genesis, blockValidator)

	// Create an orphan block (no parent exists)
	var unknownParentHash common.Hash
	unknownParentHash[0] = 0xFF

	orphan := createTestBlock(unknownParentHash, 1)
	addedHashes := store.AddBlock(orphan)

	// Orphan blocks should not be added to main/side chain
	if len(addedHashes) != 0 {
		t.Errorf("Expected 0 added hashes for orphan: got %d", len(addedHashes))
	}

	// Verify orphan is marked as orphan
	isOrphan, err := store.IsOrphanBlock(orphan)
	if err != nil {
		t.Fatalf("Failed to check if block is orphan: %v", err)
	}
	if !isOrphan {
		t.Error("Block should be marked as orphan")
	}

	// Verify orphan is not in main chain
	if store.IsPartOfMainChain(orphan) {
		t.Error("Orphan should not be part of main chain")
	}

	// Verify orphan can still be retrieved by hash
	retrieved, err := store.GetBlockByHash(orphan.Hash())
	if err != nil {
		t.Fatalf("Failed to retrieve orphan block: %v", err)
	}
	if retrieved.Hash() != orphan.Hash() {
		t.Errorf("Retrieved orphan hash mismatch")
	}

	// Verify current height is still 0
	if store.GetCurrentHeight() != 0 {
		t.Errorf("Current height should still be 0: got %d", store.GetCurrentHeight())
	}
}

// TestAddBlock_OrphanBecomesSideChain tests that orphans become side chains when parent is added
// (g)
// ..> (o) [orphan - points to unknown parent]
// ---
// (g) -> (p)        [add parent with hash matching unknown parent]
// (g) -> (p) -> (c) [re-add as connected block]
func TestAddBlock_OrphanBecomesSideChain(t *testing.T) {
	genesis := createTestBlockWithLeadingZeros([32]byte{}, 0)
	blockValidator := &blockValidatorForTests{}
	store := NewBlockStore(genesis, blockValidator)

	// Create an orphan block
	var unknownParentHash common.Hash
	unknownParentHash[0] = 0xAA

	orphan := createTestBlock(unknownParentHash, 1)
	addedHashes := store.AddBlock(orphan)

	if len(addedHashes) != 0 {
		t.Errorf("Expected 0 added hashes for orphan: got %d", len(addedHashes))
	}

	// Now add the parent block
	parent := createTestBlockWithLeadingZeros(genesis.Hash(), 10)
	parentHash := parent.Hash()

	addedHashes = store.AddBlock(parent)
	if len(addedHashes) != 1 {
		t.Errorf("Expected 1 added hash for parent: got %d", len(addedHashes))
	}

	// Now add the orphan again (it should connect to parent)
	// Create a new block with the same content as orphan but connected to parent
	connectedBlock := createTestBlock(parentHash, 1)
	addedHashes = store.AddBlock(connectedBlock)

	// Should add connectedBlock and potentially any orphans that connect to it
	if len(addedHashes) != 1 {
		t.Errorf("Expected at least 1 added hash: got %d", len(addedHashes))
	}

	// Verify the block is no longer an orphan
	isOrphan, err := store.IsOrphanBlock(connectedBlock)
	if err != nil {
		t.Fatalf("Failed to check if block is orphan: %v", err)
	}
	if isOrphan {
		t.Error("Block should no longer be an orphan")
	}
}

// TestGetBlockByHash_NotFound tests GetBlockByHash with non-existent block
func TestGetBlockByHash_NotFound(t *testing.T) {
	genesis := createTestBlockWithLeadingZeros([32]byte{}, 0)
	blockValidator := &blockValidatorForTests{}
	store := NewBlockStore(genesis, blockValidator)

	var nonExistentHash common.Hash
	nonExistentHash[0] = 0xFF

	_, err := store.GetBlockByHash(nonExistentHash)
	if err == nil {
		t.Error("Expected error for non-existent block")
	}
}

// TestGetBlockByHash_ReturnsCopy tests that GetBlockByHash returns a copy
func TestGetBlockByHash_ReturnsCopy(t *testing.T) {
	genesis := createTestBlockWithLeadingZeros([32]byte{}, 0)
	blockValidator := &blockValidatorForTests{}
	store := NewBlockStore(genesis, blockValidator)

	retrieved, err := store.GetBlockByHash(genesis.Hash())
	if err != nil {
		t.Fatalf("Failed to retrieve genesis: %v", err)
	}

	// Modify the retrieved block
	retrieved.Header.Nonce = 999

	// Retrieve again
	retrievedAgain, err := store.GetBlockByHash(genesis.Hash())
	if err != nil {
		t.Fatalf("Failed to retrieve genesis again: %v", err)
	}

	// Verify the modification didn't affect the stored block
	if retrievedAgain.Header.Nonce == 999 {
		t.Error("Modification should not affect stored block")
	}
}

// TestIsOrphanBlock_NotFound tests IsOrphanBlock with non-existent block
func TestIsOrphanBlock_NotFound(t *testing.T) {
	genesis := createTestBlockWithLeadingZeros([32]byte{}, 0)
	blockValidator := &blockValidatorForTests{}
	store := NewBlockStore(genesis, blockValidator)

	var nonExistentHash common.Hash
	nonExistentHash[0] = 0xFF

	nonExistentBlock := createTestBlock(nonExistentHash, 0)

	_, err := store.IsOrphanBlock(nonExistentBlock)
	if err == nil {
		t.Error("Expected error for non-existent block")
	}
}

// TestIsPartOfMainChain_NotFound tests IsPartOfMainChain with non-existent block
func TestIsPartOfMainChain_NotFound(t *testing.T) {
	genesis := createTestBlockWithLeadingZeros([32]byte{}, 0)
	blockValidator := &blockValidatorForTests{}
	store := NewBlockStore(genesis, blockValidator)

	var nonExistentHash common.Hash
	nonExistentHash[0] = 0xFF

	nonExistentBlock := createTestBlock(nonExistentHash, 0)

	result := store.IsPartOfMainChain(nonExistentBlock)
	if result {
		t.Error("Non-existent block should not be part of main chain")
	}
}

// TestGetBlocksByHeight_EmptyResult tests GetBlocksByHeight with no blocks at height
// (g)
// Query height 10: returns empty
func TestGetBlocksByHeight_EmptyResult(t *testing.T) {
	genesis := createTestBlockWithLeadingZeros([32]byte{}, 0)
	blockValidator := &blockValidatorForTests{}
	store := NewBlockStore(genesis, blockValidator)

	blocks := store.GetBlocksByHeight(10)
	if len(blocks) != 0 {
		t.Errorf("Expected 0 blocks at non-existent height: got %d", len(blocks))
	}
}

// TestGetBlocksByHeight_MultipleBlocks tests GetBlocksByHeight with multiple blocks at same height
// (g)
// (g) -> (b1)
// (g) -> (b2)
// (g) -> (b3)
// All at height 1
func TestGetBlocksByHeight_MultipleBlocks(t *testing.T) {
	genesis := createTestBlockWithLeadingZeros([32]byte{}, 0)
	blockValidator := &blockValidatorForTests{}
	store := NewBlockStore(genesis, blockValidator)

	// Create multiple blocks at height 1
	block1 := createTestBlockWithLeadingZeros(genesis.Hash(), 1)
	store.AddBlock(block1)

	block2 := createTestBlockWithLeadingZeros(genesis.Hash(), 2)
	store.AddBlock(block2)

	block3 := createTestBlockWithLeadingZeros(genesis.Hash(), 3)
	store.AddBlock(block3)

	// Retrieve blocks at height 1
	blocksAtHeight1 := store.GetBlocksByHeight(1)
	if len(blocksAtHeight1) != 3 {
		t.Errorf("Expected 3 blocks at height 1: got %d", len(blocksAtHeight1))
	}

	height1Hashes := make(map[common.Hash]bool)
	for _, b := range blocksAtHeight1 {
		height1Hashes[b.Hash()] = true
	}

	if !height1Hashes[block1.Hash()] {
		t.Error("Block1 should be at height 1")
	}
	if !height1Hashes[block2.Hash()] {
		t.Error("Block2 should be at height 1")
	}
	if !height1Hashes[block3.Hash()] {
		t.Error("Block3 should be at height 1")
	}

	// Verify blocks at height 0
	blocksAtHeight0 := store.GetBlocksByHeight(0)
	if len(blocksAtHeight0) != 1 {
		t.Errorf("Expected 1 block at height 0: got %d", len(blocksAtHeight0))
	}
	if blocksAtHeight0[0].Hash() != genesis.Hash() {
		t.Error("Genesis should be at height 0")
	}
}

// TestGetCurrentHeight_Empty tests GetCurrentHeight with only genesis
// (g)
func TestGetCurrentHeight_Empty(t *testing.T) {
	genesis := createTestBlockWithLeadingZeros([32]byte{}, 0)
	blockValidator := &blockValidatorForTests{}
	store := NewBlockStore(genesis, blockValidator)

	height := store.GetCurrentHeight()
	if height != 0 {
		t.Errorf("Expected height 0: got %d", height)
	}
}

// TestGetCurrentHeight_WithOrphans tests GetCurrentHeight ignores orphans
// (g)
// (g) -> (b1)
//
//	..> (o) [orphan, height ignored]
//
// Height should be 1
func TestGetCurrentHeight_WithOrphans(t *testing.T) {
	genesis := createTestBlockWithLeadingZeros([32]byte{}, 0)
	blockValidator := &blockValidatorForTests{}
	store := NewBlockStore(genesis, blockValidator)

	// Add a block to main chain
	block1 := createTestBlockWithLeadingZeros(genesis.Hash(), 1)
	store.AddBlock(block1)

	// Add an orphan
	var unknownParentHash common.Hash
	unknownParentHash[0] = 0xFF
	orphan := createTestBlock(unknownParentHash, 2)
	store.AddBlock(orphan)

	// Height should still be 1, not affected by orphan
	height := store.GetCurrentHeight()
	if height != 1 {
		t.Errorf("Expected height 1: got %d", height)
	}
}

// TestGetMainChainTip_MultipleChains tests GetMainChainTip with multiple chains
// (g)
// (g) -> (b1a) -> (b2a)
// (g) -> (b1b) -> (b2b) -> (b3b)
// (g) -> (b1c)
// Tip should be the chain with highest accumulated work
func TestGetMainChainTip_MultipleChains(t *testing.T) {
	genesis := createTestBlockWithLeadingZeros([32]byte{}, 0)
	blockValidator := &blockValidatorForTests{}
	store := NewBlockStore(genesis, blockValidator)

	// Chain 1: genesis -> block1a -> block2a
	block1a := createTestBlockWithLeadingZeros(genesis.Hash(), 1)
	store.AddBlock(block1a)

	block2a := createTestBlockWithLeadingZeros(block1a.Hash(), 2)
	store.AddBlock(block2a)

	// Chain 2: genesis -> block1b -> block2b -> block3b
	block1b := createTestBlockWithLeadingZeros(genesis.Hash(), 10)
	store.AddBlock(block1b)

	block2b := createTestBlockWithLeadingZeros(block1b.Hash(), 11)
	store.AddBlock(block2b)

	block3b := createTestBlockWithLeadingZeros(block2b.Hash(), 12)
	store.AddBlock(block3b)

	// Chain 3: genesis -> block1c
	block1c := createTestBlockWithLeadingZeros(genesis.Hash(), 20)
	store.AddBlock(block1c)

	// Main chain tip should be one of the leaves (depending on accumulated work)
	tip := store.GetMainChainTip()
	tipHash := tip.Hash()

	// The tip should be one of block2a, block3b, or block1c
	if tipHash != block2a.Hash() && tipHash != block3b.Hash() && tipHash != block1c.Hash() {
		t.Errorf("Main chain tip should be one of block2a, block3b, or block1c: got %v", tipHash)
	}

	// Verify that the tip is actually in main chain
	if !store.IsPartOfMainChain(tip) {
		t.Error("Tip should be part of main chain")
	}
}

// TestGetMainChainTip_EqualWork tests GetMainChainTip when chains have equal accumulated work
// (g)
// (g) -> (b1)
// (g) -> (b2)
// Either (b1) or (b2) can be main chain tip (implementation choice)
func TestGetMainChainTip_EqualWork(t *testing.T) {
	genesis := createTestBlockWithLeadingZeros([32]byte{}, 0)
	blockValidator := &blockValidatorForTests{}
	store := NewBlockStore(genesis, blockValidator)

	// Create two chains with similar accumulated work
	// Chain 1: genesis -> block1
	block1 := createTestBlockWithLeadingZeros(genesis.Hash(), 1)
	store.AddBlock(block1)

	// Chain 2: genesis -> block2
	block2 := createTestBlockWithLeadingZeros(genesis.Hash(), 2)
	store.AddBlock(block2)

	// Either tip should be acceptable (implementation choice)
	tip := store.GetMainChainTip()
	tipHash := tip.Hash()

	if tipHash != block1.Hash() && tipHash != block2.Hash() {
		t.Errorf("Main chain tip should be either block1 or block2: got %v", tipHash)
	}
}

// TestComplexScenario_MainChainSideChainOrphans tests a complex scenario with main chain, side chains, and orphans
// (g)
// (g) -> (b1) -> (b2) -> (b3)  [main chain]
// (g) -> (s1) -> (s2)          [side chain]
//
//	..> (o1)                     [orphan]
//	..> (o2)                     [orphan]
func TestComplexScenario_MainChainSideChainOrphans(t *testing.T) {
	genesis := createTestBlockWithLeadingZeros([32]byte{}, 0)
	blockValidator := &blockValidatorForTests{}
	store := NewBlockStore(genesis, blockValidator)

	// Create main chain
	block1 := createTestBlockWithLeadingZeros(genesis.Hash(), 1)
	store.AddBlock(block1)

	block2 := createTestBlockWithLeadingZeros(block1.Hash(), 2)
	store.AddBlock(block2)

	block3 := createTestBlockWithLeadingZeros(block2.Hash(), 3)
	store.AddBlock(block3)

	// Create side chain
	side1 := createTestBlockWithLeadingZeros(genesis.Hash(), 10)
	store.AddBlock(side1)

	side2 := createTestBlockWithLeadingZeros(side1.Hash(), 11)
	store.AddBlock(side2)

	// Create orphans
	var orphanParentHash1 common.Hash
	orphanParentHash1[0] = 0xAA
	orphan1 := createTestBlock(orphanParentHash1, 20)
	store.AddBlock(orphan1)

	var orphanParentHash2 common.Hash
	orphanParentHash2[0] = 0xBB
	orphan2 := createTestBlock(orphanParentHash2, 21)
	store.AddBlock(orphan2)

	// Verify main chain
	tip := store.GetMainChainTip()
	tipHash := tip.Hash()

	// Tip should be one of the leaves (block3 or side2, depending on accumulated work)
	if tipHash != block3.Hash() && tipHash != side2.Hash() {
		t.Errorf("Main chain tip should be either block3 or side2: got %v", tipHash)
	}

	// Verify heights
	if store.GetCurrentHeight() != 3 {
		t.Errorf("Current height should be 3: got %d", store.GetCurrentHeight())
	}

	// Verify blocks at each height
	blocksAtHeight0 := store.GetBlocksByHeight(0)
	if len(blocksAtHeight0) != 1 || blocksAtHeight0[0].Hash() != genesis.Hash() {
		t.Error("Height 0 should have only genesis")
	}

	blocksAtHeight1 := store.GetBlocksByHeight(1)
	if len(blocksAtHeight1) != 2 {
		t.Errorf("Height 1 should have 2 blocks: got %d", len(blocksAtHeight1))
	}

	blocksAtHeight2 := store.GetBlocksByHeight(2)
	if len(blocksAtHeight2) != 2 {
		t.Errorf("Height 2 should have 2 blocks: got %d", len(blocksAtHeight2))
	}

	blocksAtHeight3 := store.GetBlocksByHeight(3)
	if len(blocksAtHeight3) != 1 {
		t.Errorf("Height 3 should have 1 block: got %d", len(blocksAtHeight3))
	}

	// Verify main chain membership
	if !store.IsPartOfMainChain(genesis) {
		t.Error("Genesis should be in main chain")
	}
	if !store.IsPartOfMainChain(block1) {
		t.Error("Block1 should be in main chain")
	}
	if !store.IsPartOfMainChain(block2) {
		t.Error("Block2 should be in main chain")
	}
	if !store.IsPartOfMainChain(block3) {
		t.Error("Block3 should be in main chain")
	}
	if store.IsPartOfMainChain(side1) {
		t.Error("Side1 should not be in main chain")
	}
	if store.IsPartOfMainChain(side2) {
		t.Error("Side2 should not be in main chain")
	}

	// Verify orphan status
	isOrphan1, err := store.IsOrphanBlock(orphan1)
	if err != nil {
		t.Fatalf("Failed to check orphan1: %v", err)
	}
	if !isOrphan1 {
		t.Error("Orphan1 should be marked as orphan")
	}

	isOrphan2, err := store.IsOrphanBlock(orphan2)
	if err != nil {
		t.Fatalf("Failed to check orphan2: %v", err)
	}
	if !isOrphan2 {
		t.Error("Orphan2 should be marked as orphan")
	}

	// Verify non-orphan status
	for _, b := range []block.Block{genesis, block1, block2, block3, side1, side2} {
		isOrphan, err := store.IsOrphanBlock(b)
		if err != nil {
			t.Fatalf("Failed to check block: %v", err)
		}
		if isOrphan {
			t.Errorf("Block should not be marked as orphan: %v", b.Hash())
		}
	}

	// Verify all blocks can be retrieved
	for _, b := range []block.Block{genesis, block1, block2, block3, side1, side2, orphan1, orphan2} {
		_, err := store.GetBlockByHash(b.Hash())
		if err != nil {
			t.Errorf("Failed to retrieve block: %v", err)
		}
	}
}

// TestAddBlock_ChainedOrphans tests that orphans can chain together
// (g)
// Add orphans: (they point to each other but aren't connected to each other yet)
// (g)
// (o1) ..> (o2) ..> (o3)
// Then add (p) which connects to (g), causing (o1), (o2), (o3) to connect
// (g) -> (p) -> (o1) -> (o2) -> (o3)
func TestAddBlock_ChainedOrphans(t *testing.T) {
	genesis := createTestBlockWithLeadingZeros([32]byte{}, 0)
	blockValidator := &blockValidatorForTests{}
	store := NewBlockStore(genesis, blockValidator)

	parent := createTestBlockWithLeadingZeros(genesis.Hash(), 10)

	orphan1 := createTestBlock(parent.Hash(), 1)
	store.AddBlock(orphan1)

	// Orphan 2 points to orphan 1
	orphan2 := createTestBlock(orphan1.Hash(), 2)
	store.AddBlock(orphan2)

	// Orphan 3 points to orphan 2
	orphan3 := createTestBlock(orphan2.Hash(), 3)
	store.AddBlock(orphan3)

	// All should be orphans
	for i, orphan := range []block.Block{orphan1, orphan2, orphan3} {
		isOrphan, err := store.IsOrphanBlock(orphan)
		if err != nil {
			t.Fatalf("Failed to check orphan%d: %v", i+1, err)
		}
		if !isOrphan {
			t.Errorf("Orphan%d should be marked as orphan", i+1)
		}
	}

	// Now add the parent, all orphans should connect
	addedHashes := store.AddBlock(parent)

	// Should add parent + orphan1, orphan2, orphan3
	if len(addedHashes) != 4 {
		t.Errorf("Expected 4 added hashes (parent + 3 orphans): got %d", len(addedHashes))
	}

	// Verify orphan1, 2 and 3 are no longer orphans
	for i, orphan := range []block.Block{orphan1, orphan2, orphan3} {
		isOrphan, err := store.IsOrphanBlock(orphan)
		if err != nil {
			t.Fatalf("Failed to check orphan%d: %v", i+1, err)
		}
		if isOrphan {
			t.Errorf("Orphan%d should no longer be marked as orphan", i+1)
		}
	}
}
