package utxo_tests

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/utxo"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/utxopool"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockBlockStore implements blockchain.BlockStoreAPI for testing
type mockBlockStore struct {
	blocks       map[common.Hash]block.Block
	mainChainTip common.Hash
	heights      map[common.Hash]uint64
	mainChain    map[common.Hash]bool
}

func newMockBlockStore() *mockBlockStore {
	return &mockBlockStore{
		blocks:    make(map[common.Hash]block.Block),
		heights:   make(map[common.Hash]uint64),
		mainChain: make(map[common.Hash]bool),
	}
}

func (m *mockBlockStore) AddBlock(blk block.Block) []common.Hash {
	hash := blk.Hash()
	m.blocks[hash] = blk
	return []common.Hash{hash}
}

func (m *mockBlockStore) IsOrphanBlock(blk block.Block) (bool, error) {
	return false, nil
}

func (m *mockBlockStore) IsPartOfMainChain(blk block.Block) bool {
	return m.mainChain[blk.Hash()]
}

func (m *mockBlockStore) GetBlockByHash(hash common.Hash) (block.Block, error) {
	blk, ok := m.blocks[hash]
	if !ok {
		return block.Block{}, utxo.ErrBlockNotFound
	}
	return blk, nil
}

func (m *mockBlockStore) GetBlocksByHeight(height uint64) []block.Block {
	var result []block.Block
	for hash, h := range m.heights {
		if h == height {
			if blk, ok := m.blocks[hash]; ok {
				result = append(result, blk)
			}
		}
	}
	return result
}

func (m *mockBlockStore) GetCurrentHeight() uint64 {
	var maxHeight uint64
	for _, h := range m.heights {
		if h > maxHeight {
			maxHeight = h
		}
	}
	return maxHeight
}

func (m *mockBlockStore) GetMainChainTip() block.Block {
	return m.blocks[m.mainChainTip]
}

func (m *mockBlockStore) SetMainChainTip(hash common.Hash) {
	m.mainChainTip = hash
	m.mainChain[hash] = true
}

func (m *mockBlockStore) SetHeight(hash common.Hash, height uint64) {
	m.heights[hash] = height
}

func (m *mockBlockStore) AddToMainChain(hash common.Hash) {
	m.mainChain[hash] = true
}

// createTestBlock creates a test block with given parameters
func createTestBlock(prevHash common.Hash, txs []transaction.Transaction) block.Block {
	return block.Block{
		Header: block.BlockHeader{
			PreviousBlockHash: prevHash,
			Timestamp:         1234567890,
			Nonce:             0,
			DifficultyTarget:  1,
		},
		Transactions: txs,
	}
}

// createTestBlockWithNonce creates a test block with a specific nonce for unique hashes
func createTestBlockWithNonce(prevHash common.Hash, txs []transaction.Transaction, nonce uint32) block.Block {
	return block.Block{
		Header: block.BlockHeader{
			PreviousBlockHash: prevHash,
			Timestamp:         1234567890,
			Nonce:             nonce,
			DifficultyTarget:  1,
		},
		Transactions: txs,
	}
}

// createCoinbaseTx creates a coinbase transaction for testing
func createCoinbaseTx(value uint64, outputIndex byte) transaction.Transaction {
	var pubKeyHash transaction.PubKeyHash
	pubKeyHash[0] = outputIndex
	return transaction.Transaction{
		Inputs: []transaction.Input{},
		Outputs: []transaction.Output{
			{Value: value, PubKeyHash: pubKeyHash},
		},
	}
}

// TestMultiChainUTXOService_CreateViewAtMainChainTip tests view creation for main chain
func TestMultiChainUTXOService_CreateViewAtMainChainTip(t *testing.T) {
	// Setup
	mempool := utxo.NewMemUTXOPoolService()
	chainstate, err := utxo.NewChainStateService(
		utxo.ChainStateConfig{CacheSize: 100},
		newMockEntryDAO(),
	)
	require.NoError(t, err)
	fullNode := utxo.NewFullNodeUTXOService(mempool, chainstate)

	blockStore := newMockBlockStore()
	genesis := createTestBlock(common.Hash{}, []transaction.Transaction{createCoinbaseTx(5000, 0)})
	blockStore.AddBlock(genesis)
	blockStore.SetMainChainTip(genesis.Hash())

	service := utxo.NewMultiChainUTXOService(fullNode, blockStore)

	// Add a UTXO to chainstate
	txID := createTestTxID(1)
	outpoint := utxopool.NewOutpoint(txID, 0)
	entry := utxopool.NewUTXOEntry(createTestOutput(1000), 10, false)
	err = chainstate.Add(outpoint, entry)
	require.NoError(t, err)

	// Create view at main chain tip
	view, err := service.CreateViewAtTip(genesis.Hash())
	require.NoError(t, err)

	// Should be able to access the UTXO through the view
	got, err := view.Get(outpoint)
	require.NoError(t, err)
	assert.Equal(t, uint64(1000), got.Output.Value)
}

// TestMultiChainUTXOService_ValidateAndApplySideChainBlock tests side chain validation
func TestMultiChainUTXOService_ValidateAndApplySideChainBlock(t *testing.T) {
	// Setup
	mempool := utxo.NewMemUTXOPoolService()
	chainstate, err := utxo.NewChainStateService(
		utxo.ChainStateConfig{CacheSize: 100},
		newMockEntryDAO(),
	)
	require.NoError(t, err)
	fullNode := utxo.NewFullNodeUTXOService(mempool, chainstate)

	blockStore := newMockBlockStore()
	genesis := createTestBlock(common.Hash{}, []transaction.Transaction{createCoinbaseTx(5000, 0)})
	blockStore.AddBlock(genesis)
	blockStore.SetMainChainTip(genesis.Hash())
	blockStore.AddToMainChain(genesis.Hash())

	service := utxo.NewMultiChainUTXOService(fullNode, blockStore)

	// Add genesis coinbase UTXO to chainstate
	genesisTx := genesis.Transactions[0]
	genesisTxID := genesisTx.TransactionId()
	genesisOutpoint := utxopool.NewOutpoint(genesisTxID, 0)
	err = chainstate.Add(genesisOutpoint, utxopool.NewUTXOEntry(genesisTx.Outputs[0], 0, true))
	require.NoError(t, err)

	// Create a side chain block that spends the genesis coinbase
	sideChainTx := transaction.Transaction{
		Inputs:  []transaction.Input{{PrevTxID: genesisTxID, OutputIndex: 0}},
		Outputs: []transaction.Output{createTestOutput(4500)},
	}
	sideChainBlock := createTestBlock(genesis.Hash(), []transaction.Transaction{
		createCoinbaseTx(50, 1), // coinbase
		sideChainTx,
	})
	blockStore.AddBlock(sideChainBlock)

	// Validate and apply the side chain block
	delta, err := service.ValidateAndApplySideChainBlock(sideChainBlock, 1)
	require.NoError(t, err)

	// Delta should have the correct fork point
	assert.Equal(t, genesis.Hash(), delta.ForkPoint)
	assert.Equal(t, sideChainBlock.Hash(), delta.ChainTip)

	// Should have spent the genesis UTXO
	assert.Len(t, delta.SpentUTXOs, 1)

	// Should have added new UTXOs (coinbase + tx output)
	assert.Len(t, delta.AddedUTXOs, 2)
}

// TestMultiChainUTXOService_SideChainDoesNotAffectMainChain verifies main chain isolation
func TestMultiChainUTXOService_SideChainDoesNotAffectMainChain(t *testing.T) {
	// Setup
	mempool := utxo.NewMemUTXOPoolService()
	chainstate, err := utxo.NewChainStateService(
		utxo.ChainStateConfig{CacheSize: 100},
		newMockEntryDAO(),
	)
	require.NoError(t, err)
	fullNode := utxo.NewFullNodeUTXOService(mempool, chainstate)

	blockStore := newMockBlockStore()
	genesis := createTestBlock(common.Hash{}, []transaction.Transaction{createCoinbaseTx(5000, 0)})
	blockStore.AddBlock(genesis)
	blockStore.SetMainChainTip(genesis.Hash())
	blockStore.AddToMainChain(genesis.Hash())

	service := utxo.NewMultiChainUTXOService(fullNode, blockStore)

	// Add a UTXO to chainstate
	txID := createTestTxID(1)
	outpoint := utxopool.NewOutpoint(txID, 0)
	entry := utxopool.NewUTXOEntry(createTestOutput(1000), 10, false)
	err = chainstate.Add(outpoint, entry)
	require.NoError(t, err)

	// Create a side chain block that spends the UTXO
	sideChainTx := transaction.Transaction{
		Inputs:  []transaction.Input{{PrevTxID: txID, OutputIndex: 0}},
		Outputs: []transaction.Output{createTestOutput(900)},
	}
	sideChainBlock := createTestBlock(genesis.Hash(), []transaction.Transaction{
		createCoinbaseTx(50, 1),
		sideChainTx,
	})
	blockStore.AddBlock(sideChainBlock)

	// Apply side chain block
	_, err = service.ValidateAndApplySideChainBlock(sideChainBlock, 1)
	require.NoError(t, err)

	// The main chain UTXO should still exist
	assert.True(t, chainstate.Contains(outpoint))
	got, err := chainstate.Get(outpoint)
	require.NoError(t, err)
	assert.Equal(t, uint64(1000), got.Output.Value)
}

// TestMultiChainUTXOService_ExtendSideChain tests extending an existing side chain
func TestMultiChainUTXOService_ExtendSideChain(t *testing.T) {
	// Setup
	mempool := utxo.NewMemUTXOPoolService()
	chainstate, err := utxo.NewChainStateService(
		utxo.ChainStateConfig{CacheSize: 100},
		newMockEntryDAO(),
	)
	require.NoError(t, err)
	fullNode := utxo.NewFullNodeUTXOService(mempool, chainstate)

	blockStore := newMockBlockStore()
	genesis := createTestBlock(common.Hash{}, []transaction.Transaction{createCoinbaseTx(5000, 0)})
	blockStore.AddBlock(genesis)
	blockStore.SetMainChainTip(genesis.Hash())
	blockStore.AddToMainChain(genesis.Hash())

	service := utxo.NewMultiChainUTXOService(fullNode, blockStore)

	// Add genesis coinbase UTXO
	genesisTx := genesis.Transactions[0]
	genesisTxID := genesisTx.TransactionId()
	err = chainstate.Add(
		utxopool.NewOutpoint(genesisTxID, 0),
		utxopool.NewUTXOEntry(genesisTx.Outputs[0], 0, true),
	)
	require.NoError(t, err)

	// First side chain block
	sideBlock1 := createTestBlock(genesis.Hash(), []transaction.Transaction{
		createCoinbaseTx(50, 1),
	})
	blockStore.AddBlock(sideBlock1)

	delta1, err := service.ValidateAndApplySideChainBlock(sideBlock1, 1)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), delta1.BlockCount)

	// Get the coinbase txID from first side block
	sideBlock1CoinbaseTxID := sideBlock1.Transactions[0].TransactionId()

	// Second side chain block - spends first block's coinbase
	sideBlock2Tx := transaction.Transaction{
		Inputs:  []transaction.Input{{PrevTxID: sideBlock1CoinbaseTxID, OutputIndex: 0}},
		Outputs: []transaction.Output{createTestOutput(40)},
	}
	sideBlock2 := createTestBlock(sideBlock1.Hash(), []transaction.Transaction{
		createCoinbaseTx(50, 2),
		sideBlock2Tx,
	})
	blockStore.AddBlock(sideBlock2)

	delta2, err := service.ValidateAndApplySideChainBlock(sideBlock2, 2)
	require.NoError(t, err)

	// Should have extended the chain
	assert.Equal(t, uint64(2), delta2.BlockCount)
	assert.Equal(t, genesis.Hash(), delta2.ForkPoint)
	assert.Equal(t, sideBlock2.Hash(), delta2.ChainTip)
}

// mockEntryDAO implements utxopool.EntryDAO for testing
type mockEntryDAO struct {
	entries map[string]utxopool.UTXOEntry
	index   map[transaction.PubKeyHash][]utxopool.Outpoint
}

func newMockEntryDAO() *mockEntryDAO {
	return &mockEntryDAO{
		entries: make(map[string]utxopool.UTXOEntry),
		index:   make(map[transaction.PubKeyHash][]utxopool.Outpoint),
	}
}

func (m *mockEntryDAO) Find(outpoint utxopool.Outpoint) (utxopool.UTXOEntry, error) {
	key := string(outpoint.Key())
	if entry, ok := m.entries[key]; ok {
		return entry, nil
	}
	return utxopool.UTXOEntry{}, utxo.ErrUTXONotFound
}

func (m *mockEntryDAO) Update(outpoint utxopool.Outpoint, entry utxopool.UTXOEntry) error {
	key := string(outpoint.Key())
	m.entries[key] = entry
	m.index[entry.Output.PubKeyHash] = append(m.index[entry.Output.PubKeyHash], outpoint)
	return nil
}

func (m *mockEntryDAO) Delete(outpoint utxopool.Outpoint) error {
	key := string(outpoint.Key())
	delete(m.entries, key)
	return nil
}

func (m *mockEntryDAO) FindByPubKeyHash(pubKeyHash transaction.PubKeyHash) ([]utxopool.Outpoint, error) {
	return m.index[pubKeyHash], nil
}

func (m *mockEntryDAO) Close() error {
	return nil
}

func (m *mockEntryDAO) Persist() error {
	return nil
}

// TestSideChainDelta_ParentTip tests that ParentTip is tracked correctly
func TestSideChainDelta_ParentTip(t *testing.T) {
	forkPoint := createTestHash(1)
	chainTip := createTestHash(2)
	delta := utxo.NewSideChainDelta(forkPoint, chainTip)

	// Initially, ParentTip should equal ForkPoint
	assert.Equal(t, forkPoint, delta.ParentTip)

	// After applying a view, ParentTip should be the previous ChainTip
	base := newMockBaseProvider()
	view := utxo.NewEphemeralUTXOView(base)
	tx := &transaction.Transaction{
		Inputs:  []transaction.Input{},
		Outputs: []transaction.Output{createTestOutput(1000)},
	}
	err := view.ApplyTx(tx, createTestTxID(1), 10, true)
	require.NoError(t, err)

	newTip := createTestHash(3)
	delta.ApplyView(view, newTip)

	assert.Equal(t, chainTip, delta.ParentTip)
	assert.Equal(t, newTip, delta.ChainTip)
}

// TestSideChainDeltaStore_ParentIndex tests the parent index functionality
func TestSideChainDeltaStore_ParentIndex(t *testing.T) {
	store := utxo.NewSideChainDeltaStore()

	forkPoint := createTestHash(1)
	chainTip1 := createTestHash(2)
	chainTip2 := createTestHash(3)

	// Create two deltas with the same fork point but different tips
	delta1 := utxo.NewSideChainDelta(forkPoint, chainTip1)
	delta2 := utxo.NewSideChainDeltaWithParent(forkPoint, chainTip2, chainTip1)

	store.Put(delta1)
	store.Put(delta2)

	// Check GetByParent
	children := store.GetByParent(chainTip1)
	assert.Len(t, children, 1)
	assert.Equal(t, chainTip2, children[0].ChainTip)

	// Check GetByForkPoint
	tips := store.GetByForkPoint(forkPoint)
	assert.Len(t, tips, 2)
}

// TestBuildMainChainDelta tests building a delta by replaying main chain blocks
func TestBuildMainChainDelta(t *testing.T) {
	mempool := utxo.NewMemUTXOPoolService()
	chainstate, err := utxo.NewChainStateService(
		utxo.ChainStateConfig{CacheSize: 100},
		newMockEntryDAO(),
	)
	require.NoError(t, err)
	fullNode := utxo.NewFullNodeUTXOService(mempool, chainstate)

	blockStore := newMockBlockStore()

	// Create genesis block
	genesis := createTestBlock(common.Hash{}, []transaction.Transaction{createCoinbaseTx(5000, 0)})
	blockStore.AddBlock(genesis)
	blockStore.SetMainChainTip(genesis.Hash())
	blockStore.AddToMainChain(genesis.Hash())
	blockStore.SetHeight(genesis.Hash(), 0)

	// Create block 1
	block1 := createTestBlockWithNonce(genesis.Hash(), []transaction.Transaction{createCoinbaseTx(50, 1)}, 1)
	blockStore.AddBlock(block1)
	blockStore.AddToMainChain(block1.Hash())
	blockStore.SetHeight(block1.Hash(), 1)
	blockStore.SetMainChainTip(block1.Hash())

	service := utxo.NewMultiChainUTXOService(fullNode, blockStore)

	// Create view at main chain tip - should include replayed blocks
	view, err := service.CreateViewAtTip(block1.Hash())
	require.NoError(t, err)
	require.NotNil(t, view)

	// The view should be able to see UTXOs from the replayed blocks
	// Verify the coinbase UTXO from block 1 is accessible
	block1TxID := block1.Transactions[0].TransactionId()
	outpoint := utxopool.NewOutpoint(block1TxID, 0)
	entry, err := view.Get(outpoint)
	require.NoError(t, err)
	assert.Equal(t, uint64(50), entry.Output.Value)
}

// TestChainedDeltaBaseProvider tests stacking multiple deltas
func TestChainedDeltaBaseProvider(t *testing.T) {
	base := newMockBaseProvider()

	// Add a UTXO to base
	txID1 := createTestTxID(1)
	outpoint1 := utxopool.NewOutpoint(txID1, 0)
	base.Add(outpoint1, utxopool.NewUTXOEntry(createTestOutput(1000), 10, false))

	// First delta: adds a UTXO
	delta1 := utxo.NewSideChainDelta(createTestHash(1), createTestHash(2))
	txID2 := createTestTxID(2)
	outpoint2 := utxopool.NewOutpoint(txID2, 0)
	delta1.AddedUTXOs[string(outpoint2.Key())] = utxopool.NewUTXOEntry(createTestOutput(500), 11, false)

	// Second delta: spends the UTXO added in delta1, adds new UTXO
	delta2 := utxo.NewSideChainDeltaWithParent(createTestHash(1), createTestHash(3), createTestHash(2))
	delta2.SpentUTXOs[string(outpoint2.Key())] = struct{}{}
	txID3 := createTestTxID(3)
	outpoint3 := utxopool.NewOutpoint(txID3, 0)
	delta2.AddedUTXOs[string(outpoint3.Key())] = utxopool.NewUTXOEntry(createTestOutput(400), 12, false)

	// Create chained provider
	provider := utxo.NewChainedDeltaBaseProvider(base, []*utxo.SideChainDelta{delta1, delta2})

	// outpoint1 should still be accessible from base
	assert.True(t, provider.Contains(outpoint1))
	got1, err := provider.Get(outpoint1)
	require.NoError(t, err)
	assert.Equal(t, uint64(1000), got1.Output.Value)

	// outpoint2 was added in delta1 but spent in delta2 - should not be accessible
	assert.False(t, provider.Contains(outpoint2))
	_, err = provider.Get(outpoint2)
	assert.ErrorIs(t, err, utxo.ErrUTXOAlreadySpent)

	// outpoint3 was added in delta2 - should be accessible
	assert.True(t, provider.Contains(outpoint3))
	got3, err := provider.Get(outpoint3)
	require.NoError(t, err)
	assert.Equal(t, uint64(400), got3.Output.Value)
}

// TestMultiChainUTXOService_BlockReplayForShallowFork tests that shallow forks use block replay
func TestMultiChainUTXOService_BlockReplayForShallowFork(t *testing.T) {
	mempool := utxo.NewMemUTXOPoolService()
	chainstate, err := utxo.NewChainStateService(
		utxo.ChainStateConfig{CacheSize: 100},
		newMockEntryDAO(),
	)
	require.NoError(t, err)
	fullNode := utxo.NewFullNodeUTXOService(mempool, chainstate)

	blockStore := newMockBlockStore()

	// Create genesis block with coinbase
	genesis := createTestBlock(common.Hash{}, []transaction.Transaction{createCoinbaseTx(5000, 0)})
	blockStore.AddBlock(genesis)
	blockStore.SetMainChainTip(genesis.Hash())
	blockStore.AddToMainChain(genesis.Hash())
	blockStore.SetHeight(genesis.Hash(), 0)

	// Add genesis coinbase to chainstate (simulating confirmed block)
	genesisTx := genesis.Transactions[0]
	genesisTxID := genesisTx.TransactionId()
	err = chainstate.Add(
		utxopool.NewOutpoint(genesisTxID, 0),
		utxopool.NewUTXOEntry(genesisTx.Outputs[0], 0, true),
	)
	require.NoError(t, err)

	// Create main chain block 1 (this will be in "unconfirmed" range)
	block1 := createTestBlockWithNonce(genesis.Hash(), []transaction.Transaction{createCoinbaseTx(50, 1)}, 1)
	blockStore.AddBlock(block1)
	blockStore.AddToMainChain(block1.Hash())
	blockStore.SetHeight(block1.Hash(), 1)
	blockStore.SetMainChainTip(block1.Hash())

	service := utxo.NewMultiChainUTXOService(fullNode, blockStore)

	// Create a side chain block that forks from genesis
	sideBlock := createTestBlockWithNonce(genesis.Hash(), []transaction.Transaction{createCoinbaseTx(60, 2)}, 2)
	blockStore.AddBlock(sideBlock)
	blockStore.SetHeight(sideBlock.Hash(), 1)

	// Validate side chain block - this should work using chainstate, not mempool
	delta, err := service.ValidateAndApplySideChainBlock(sideBlock, 1)
	require.NoError(t, err)
	require.NotNil(t, delta)

	// Delta should have the correct structure
	assert.Equal(t, genesis.Hash(), delta.ForkPoint)
	assert.Equal(t, sideBlock.Hash(), delta.ChainTip)
	assert.Equal(t, uint64(1), delta.BlockCount)
}

// TestMultiChainUTXOService_CanCreateViewAt tests the CanCreateViewAt helper
func TestMultiChainUTXOService_CanCreateViewAt(t *testing.T) {
	mempool := utxo.NewMemUTXOPoolService()
	chainstate, err := utxo.NewChainStateService(
		utxo.ChainStateConfig{CacheSize: 100},
		newMockEntryDAO(),
	)
	require.NoError(t, err)
	fullNode := utxo.NewFullNodeUTXOService(mempool, chainstate)

	blockStore := newMockBlockStore()
	genesis := createTestBlock(common.Hash{}, []transaction.Transaction{createCoinbaseTx(5000, 0)})
	blockStore.AddBlock(genesis)
	blockStore.SetMainChainTip(genesis.Hash())
	blockStore.AddToMainChain(genesis.Hash())

	service := utxo.NewMultiChainUTXOService(fullNode, blockStore)

	// Main chain tip should always be available
	assert.True(t, service.CanCreateViewAt(genesis.Hash()))

	// Non-existent block should not be available
	assert.False(t, service.CanCreateViewAt(createTestHash(99)))
}

// TestMultiChainUTXOService_GetAllSideChainDeltas tests retrieving all deltas
func TestMultiChainUTXOService_GetAllSideChainDeltas(t *testing.T) {
	mempool := utxo.NewMemUTXOPoolService()
	chainstate, err := utxo.NewChainStateService(
		utxo.ChainStateConfig{CacheSize: 100},
		newMockEntryDAO(),
	)
	require.NoError(t, err)
	fullNode := utxo.NewFullNodeUTXOService(mempool, chainstate)

	blockStore := newMockBlockStore()
	genesis := createTestBlock(common.Hash{}, []transaction.Transaction{createCoinbaseTx(5000, 0)})
	blockStore.AddBlock(genesis)
	blockStore.SetMainChainTip(genesis.Hash())
	blockStore.AddToMainChain(genesis.Hash())

	service := utxo.NewMultiChainUTXOService(fullNode, blockStore)

	// Add genesis coinbase to chainstate
	genesisTx := genesis.Transactions[0]
	genesisTxID := genesisTx.TransactionId()
	err = chainstate.Add(
		utxopool.NewOutpoint(genesisTxID, 0),
		utxopool.NewUTXOEntry(genesisTx.Outputs[0], 0, true),
	)
	require.NoError(t, err)

	// Create one side chain block
	sideBlock1 := createTestBlock(genesis.Hash(), []transaction.Transaction{createCoinbaseTx(50, 1)})
	blockStore.AddBlock(sideBlock1)

	delta1, err := service.ValidateAndApplySideChainBlock(sideBlock1, 1)
	require.NoError(t, err)

	// Should have one side chain delta
	deltas := service.GetAllSideChainDeltas()
	require.Len(t, deltas, 1)
	assert.Equal(t, sideBlock1.Hash(), delta1.ChainTip)
}

// TestBuildDeltaForDisconnectedBlocks tests creating a delta from disconnected blocks
func TestBuildDeltaForDisconnectedBlocks(t *testing.T) {
	mempool := utxo.NewMemUTXOPoolService()
	chainstate, err := utxo.NewChainStateService(
		utxo.ChainStateConfig{CacheSize: 100},
		newMockEntryDAO(),
	)
	require.NoError(t, err)
	fullNode := utxo.NewFullNodeUTXOService(mempool, chainstate)

	blockStore := newMockBlockStore()
	genesis := createTestBlock(common.Hash{}, []transaction.Transaction{createCoinbaseTx(5000, 0)})
	blockStore.AddBlock(genesis)
	blockStore.SetMainChainTip(genesis.Hash())
	blockStore.AddToMainChain(genesis.Hash())
	blockStore.SetHeight(genesis.Hash(), 0)

	// Add genesis coinbase to chainstate
	genesisTx := genesis.Transactions[0]
	genesisTxID := genesisTx.TransactionId()
	err = chainstate.Add(
		utxopool.NewOutpoint(genesisTxID, 0),
		utxopool.NewUTXOEntry(genesisTx.Outputs[0], 0, true),
	)
	require.NoError(t, err)

	// Add main chain block - use nonce 1 for unique hash
	mainBlock := createTestBlockWithNonce(genesis.Hash(), []transaction.Transaction{
		createCoinbaseTx(100, 1),
	}, 1)
	blockStore.AddBlock(mainBlock)
	blockStore.AddToMainChain(mainBlock.Hash())
	blockStore.SetHeight(mainBlock.Hash(), 1)
	blockStore.SetMainChainTip(mainBlock.Hash())

	// Apply main block to chainstate
	mainCoinbaseTx := mainBlock.Transactions[0]
	mainCoinbaseTxID := mainCoinbaseTx.TransactionId()
	err = chainstate.Add(
		utxopool.NewOutpoint(mainCoinbaseTxID, 0),
		utxopool.NewUTXOEntry(mainCoinbaseTx.Outputs[0], 1, true),
	)
	require.NoError(t, err)

	service := utxo.NewMultiChainUTXOService(fullNode, blockStore)

	// Create side chain that will become main chain - use nonce 2 for unique hash
	sideBlock := createTestBlockWithNonce(genesis.Hash(), []transaction.Transaction{
		createCoinbaseTx(200, 2),
	}, 2)
	blockStore.AddBlock(sideBlock)
	blockStore.SetHeight(sideBlock.Hash(), 1)

	// Verify blocks have different hashes
	require.NotEqual(t, mainBlock.Hash(), sideBlock.Hash(), "Test setup error: blocks should have different hashes")

	// Validate side chain block
	_, err = service.ValidateAndApplySideChainBlock(sideBlock, 1)
	require.NoError(t, err)

	// Simulate reorg: promote side chain, old main becomes side chain
	disconnectedBlocks := []block.Block{mainBlock}
	connectedBlocks := []block.Block{sideBlock}

	err = service.PromoteSideChainToMain(sideBlock.Hash(), genesis.Hash(), disconnectedBlocks, connectedBlocks)
	require.NoError(t, err)

	// Old main chain should now be stored as a side chain delta
	oldMainDelta, exists := service.GetSideChainDelta(mainBlock.Hash())
	require.True(t, exists, "Old main chain should be stored as side chain delta")
	require.NotNil(t, oldMainDelta)
	assert.Equal(t, genesis.Hash(), oldMainDelta.ForkPoint)
	assert.Equal(t, mainBlock.Hash(), oldMainDelta.ChainTip)
	assert.Equal(t, uint64(1), oldMainDelta.BlockCount)
}
