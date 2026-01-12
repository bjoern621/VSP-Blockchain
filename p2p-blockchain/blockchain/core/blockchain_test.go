package core

import (
	"errors"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/utxo"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/validation"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/utxopool"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/inv"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"s3b/vsp-blockchain/p2p-blockchain/miner/api/observer"
	"testing"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/stretchr/testify/assert"
)

type mockBlockchainSender struct {
	called     bool
	lastMsg    []*inv.InvVector
	lastPeerID common.PeerId
	callsCount int

	// For Block method testing
	broadcastAddedBlocksCalled   bool
	broadcastAddedBlocksHashes   []common.Hash
	broadcastAddedBlocksExcluded common.PeerId
	requestMissingHeadersCalled  bool
	requestMissingHeadersLocator block.BlockLocator
	requestMissingHeadersPeerId  common.PeerId
}

func (m *mockBlockchainSender) SendInv(inventory []*inv.InvVector, peerId common.PeerId) {
	m.called = true
	m.callsCount++
	m.lastMsg = inventory
	m.lastPeerID = peerId
}

func (m *mockBlockchainSender) SendGetData(msg []*inv.InvVector, peerId common.PeerId) {
	m.called = true
	m.callsCount++
	m.lastMsg = msg
	m.lastPeerID = peerId
}

func (m *mockBlockchainSender) BroadcastInvExclusionary(msg []*inv.InvVector, peerId common.PeerId) {}
func (m *mockBlockchainSender) BroadcastAddedBlocks(blockHashes []common.Hash, excludedPeerId common.PeerId) {
	m.broadcastAddedBlocksCalled = true
	m.broadcastAddedBlocksHashes = blockHashes
	m.broadcastAddedBlocksExcluded = excludedPeerId
}
func (m *mockBlockchainSender) RequestMissingBlockHeaders(blockLocator block.BlockLocator, peerId common.PeerId) {
	m.requestMissingHeadersCalled = true
	m.requestMissingHeadersLocator = blockLocator
	m.requestMissingHeadersPeerId = peerId
}

// mockBlockValidator is a mock for validation.BlockValidationAPI
type mockBlockValidator struct {
	sanityCheckResult    bool
	sanityCheckErr       error
	validateHeaderResult bool
	validateHeaderErr    error
	fullValidationResult bool
	fullValidationErr    error

	sanityCheckCalled    bool
	validateHeaderCalled bool
	fullValidationCalled bool
}

func (m *mockBlockValidator) SanityCheck(b block.Block) (bool, error) {
	m.sanityCheckCalled = true
	return m.sanityCheckResult, m.sanityCheckErr
}

func (m *mockBlockValidator) ValidateHeader(b block.Block) (bool, error) {
	m.validateHeaderCalled = true
	return m.validateHeaderResult, m.validateHeaderErr
}

func (m *mockBlockValidator) FullValidation(b block.Block) (bool, error) {
	m.fullValidationCalled = true
	return m.fullValidationResult, m.fullValidationErr
}

// mockBlockStore is a mock for blockchain.BlockStore
type mockBlockStore struct {
	addedBlocks           []common.Hash
	isOrphanResult        bool
	isOrphanErr           error
	currentHeight         uint64
	mainChainTip          block.Block
	isOrphanCalled        bool
	getMainChainTipCalled bool

	// For buildBlockLocator
	blocksByHeight    map[uint64][]block.Block
	isPartOfMainChain func(block.Block) bool

	// Configurable return value for AddBlock (simulating multiple blocks added, e.g., connecting orphans)
	addBlockReturnValue []common.Hash
}

func (m *mockBlockStore) GetBlockByHash(hash common.Hash) (block.Block, error) {
	return block.Block{}, nil
}

func (m *mockBlockStore) AddBlock(b block.Block) []common.Hash {
	m.addedBlocks = append(m.addedBlocks, b.Hash())
	if m.addBlockReturnValue != nil {
		return m.addBlockReturnValue
	}
	return []common.Hash{b.Hash()}
}

func (m *mockBlockStore) IsOrphanBlock(b block.Block) (bool, error) {
	m.isOrphanCalled = true
	return m.isOrphanResult, m.isOrphanErr
}

func (m *mockBlockStore) GetCurrentHeight() uint64 {
	return m.currentHeight
}

func (m *mockBlockStore) GetMainChainTip() block.Block {
	m.getMainChainTipCalled = true
	return m.mainChainTip
}

func (m *mockBlockStore) GetBlocksByHeight(height uint64) []block.Block {
	if m.blocksByHeight != nil {
		return m.blocksByHeight[height]
	}
	return []block.Block{}
}

func (m *mockBlockStore) IsPartOfMainChain(b block.Block) bool {
	if m.isPartOfMainChain != nil {
		return m.isPartOfMainChain(b)
	}
	return true
}

// mockChainReorganization is a mock for ChainReorganization
type mockChainReorganization struct {
	checkAndReorganizeResult bool
	checkAndReorganizeErr    error
	checkAndReorganizeCalled bool
	lastTipHash              common.Hash
}

func (m *mockChainReorganization) CheckAndReorganize(tipHash common.Hash) (bool, error) {
	m.checkAndReorganizeCalled = true
	m.lastTipHash = tipHash
	return m.checkAndReorganizeResult, m.checkAndReorganizeErr
}

// Helper function to create a test block
func createTestBlock(prevHash common.Hash, nonce uint32) block.Block {
	return block.Block{
		Header: block.BlockHeader{
			PreviousBlockHash: prevHash,
			MerkleRoot:        [32]byte{1, 2, 3},
			Timestamp:         1234567890,
			DifficultyTarget:  10,
			Nonce:             nonce,
		},
		Transactions: []transaction.Transaction{
			transaction.NewCoinbaseTransaction([20]byte{1, 2, 3}, 100),
		},
	}
}

type mockLookupAPIImpl struct{}

var _ utxo.LookupService = (*mockLookupAPIImpl)(nil)

func (mockLookupAPIImpl) GetUTXO(txID transaction.TransactionID, outputIndex uint32) (transaction.Output, error) {
	return transaction.Output{}, nil
}
func (mockLookupAPIImpl) GetUTXOEntry(outpoint utxopool.Outpoint) (utxopool.UTXOEntry, error) {
	return utxopool.UTXOEntry{}, nil
}
func (mockLookupAPIImpl) ContainsUTXO(outpoint utxopool.Outpoint) bool {
	return true
}
func (mockLookupAPIImpl) GetUTXOsByPubKeyHash(pubKeyHash transaction.PubKeyHash) ([]transaction.UTXO, error) {
	return []transaction.UTXO{}, nil
}

func TestBlockchain_Inv_InvokesRequestDataByCallingSendGetData(t *testing.T) {
	// Arrange: create blockchain with mocked sender
	sender := &mockBlockchainSender{}
	bc := NewBlockchain(sender, validation.NewValidationService(mockLookupAPIImpl{}), nil, nil, nil)

	peerID := common.PeerId("peer-1")

	var h common.Hash
	h[0] = 0xAB // arbitrary non-zero hash to make assertions clearer

	invVector := &inv.InvVector{
		InvType: inv.InvTypeMsgTx,
		Hash:    h,
	}
	inventory := []*inv.InvVector{
		invVector,
	}
	// Act: receive Inv message
	bc.Inv(inventory, peerID)

	// Assert: requestData() is exercised by observing its effect: SendGetData is invoked
	if !sender.called {
		t.Fatalf("expected SendGetData to be called (requestData invoked), but it was not called")
	}
	if sender.callsCount != 1 {
		t.Fatalf("expected SendGetData to be called once, but was called %d times", sender.callsCount)
	}
	if sender.lastPeerID != peerID {
		t.Fatalf("expected peerID %q, got %q", peerID, sender.lastPeerID)
	}
	if len(sender.lastMsg) != 1 {
		t.Fatalf("expected GetData inventory length 1, got %d", len(sender.lastMsg))
	}
	if sender.lastMsg[0].InvType != inv.InvTypeMsgTx {
		t.Fatalf("expected inventory[0].Type %v, got %v", inv.InvTypeMsgTx, sender.lastMsg[0].InvType)
	}
	if sender.lastMsg[0].Hash != h {
		t.Fatalf("expected inventory[0].Hash %v, got %v", h, sender.lastMsg[0].Hash)
	}
}

// =========================================================================
// Tests for Block method
// =========================================================================

// TestBlockchain_Block_SanityCheckFailure verifies that when SanityCheck fails,
// the block processing stops early and no further processing occurs.
func TestBlockchain_Block_SanityCheckFailure(t *testing.T) {
	// Arrange
	sender := &mockBlockchainSender{}
	validator := &mockBlockValidator{
		sanityCheckResult: false,
		sanityCheckErr:    errors.New("block must contain at least one transaction"),
	}
	store := &mockBlockStore{}
	reorg := &mockChainReorganization{}

	bc := &Blockchain{
		blockchainMsgSender: sender,
		blockValidator:      validator,
		blockStore:          store,
		chainReorganization: reorg,
	}

	testBlock := createTestBlock(common.Hash{}, 123)
	peerID := common.PeerId("peer-1")

	// Act
	bc.Block(testBlock, peerID)

	// Assert: SanityCheck was called
	assert.True(t, validator.sanityCheckCalled, "SanityCheck should be called")

	// Assert: No further validation occurred
	assert.False(t, validator.validateHeaderCalled, "ValidateHeader should not be called when SanityCheck fails")
	assert.False(t, validator.fullValidationCalled, "FullValidation should not be called when SanityCheck fails")
	assert.False(t, store.isOrphanCalled, "IsOrphanBlock should not be called when SanityCheck fails")
	assert.False(t, reorg.checkAndReorganizeCalled, "CheckAndReorganize should not be called when SanityCheck fails")
	assert.False(t, sender.broadcastAddedBlocksCalled, "BroadcastAddedBlocks should not be called when SanityCheck fails")
}

// TestBlockchain_Block_ValidateHeaderFailure verifies that when ValidateHeader fails,
// the block processing stops and no further processing occurs.
func TestBlockchain_Block_ValidateHeaderFailure(t *testing.T) {
	// Arrange
	sender := &mockBlockchainSender{}
	validator := &mockBlockValidator{
		sanityCheckResult:    true,
		validateHeaderResult: false,
		validateHeaderErr:    errors.New("block hash is not smaller than target"),
	}
	store := &mockBlockStore{}
	reorg := &mockChainReorganization{}

	bc := &Blockchain{
		blockchainMsgSender: sender,
		blockValidator:      validator,
		blockStore:          store,
		chainReorganization: reorg,
	}

	testBlock := createTestBlock(common.Hash{}, 123)
	peerID := common.PeerId("peer-2")

	// Act
	bc.Block(testBlock, peerID)

	// Assert: Both sanity checks were called
	assert.True(t, validator.sanityCheckCalled, "SanityCheck should be called")
	assert.True(t, validator.validateHeaderCalled, "ValidateHeader should be called")

	// Assert: No further processing occurred
	assert.False(t, validator.fullValidationCalled, "FullValidation should not be called when ValidateHeader fails")
	assert.False(t, store.isOrphanCalled, "IsOrphanBlock should not be called when ValidateHeader fails")
	assert.False(t, reorg.checkAndReorganizeCalled, "CheckAndReorganize should not be called when ValidateHeader fails")
	assert.False(t, sender.broadcastAddedBlocksCalled, "BroadcastAddedBlocks should not be called when ValidateHeader fails")
}

// TestBlockchain_Block_IsOrphanRequestsMissingHeaders verifies that when a block
// is identified as an orphan, missing block headers are requested.
func TestBlockchain_Block_IsOrphanRequestsMissingHeaders(t *testing.T) {
	// Arrange
	sender := &mockBlockchainSender{}
	validator := &mockBlockValidator{
		sanityCheckResult:    true,
		validateHeaderResult: true,
		fullValidationResult: true,
	}
	store := &mockBlockStore{
		isOrphanResult: true,
		currentHeight:  10,
		blocksByHeight: make(map[uint64][]block.Block),
	}
	reorg := &mockChainReorganization{}

	// Set up a main chain tip for buildBlockLocator
	mainTipBlock := createTestBlock(common.Hash{}, 999)
	mainTipBlock.Header.PreviousBlockHash = common.Hash{} // Genesis-like
	store.mainChainTip = mainTipBlock
	store.blocksByHeight[0] = []block.Block{mainTipBlock}
	store.isPartOfMainChain = func(b block.Block) bool {
		return true
	}

	bc := &Blockchain{
		blockchainMsgSender: sender,
		blockValidator:      validator,
		blockStore:          store,
		chainReorganization: reorg,
		mempool:             NewMempool(nil, nil),
		observers:           mapset.NewSet[observer.BlockchainObserverAPI](),
	}

	// Create a block with a non-zero parent hash (simulating an orphan)
	var parentHash common.Hash
	parentHash[0] = 0xFF
	testBlock := createTestBlock(parentHash, 123)
	peerID := common.PeerId("peer-orphan")

	// Act
	bc.Block(testBlock, peerID)

	// Assert: Initial validations were called
	assert.True(t, validator.sanityCheckCalled, "SanityCheck should be called")
	assert.True(t, validator.validateHeaderCalled, "ValidateHeader should be called")
	assert.True(t, store.isOrphanCalled, "IsOrphanBlock should be called")

	// Assert: Missing headers were requested
	assert.True(t, sender.requestMissingHeadersCalled, "RequestMissingBlockHeaders should be called for orphan blocks")
	assert.Equal(t, peerID, sender.requestMissingHeadersPeerId, "PeerID should match")

	// Assert: No further processing for orphans
	assert.False(t, validator.fullValidationCalled, "FullValidation should not be called for orphan blocks")
	assert.False(t, reorg.checkAndReorganizeCalled, "CheckAndReorganize should not be called for orphan blocks")
	assert.False(t, sender.broadcastAddedBlocksCalled, "BroadcastAddedBlocks should not be called for orphan blocks")

	// Assert: The locator includes the parent hash as first entry
	if len(sender.requestMissingHeadersLocator.BlockLocatorHashes) > 0 {
		assert.Equal(t, parentHash, sender.requestMissingHeadersLocator.BlockLocatorHashes[0],
			"First hash in locator should be the orphan parent hash")
	}
}

// TestBlockchain_Block_FullValidationFailure verifies that when FullValidation fails,
// the block is not applied to the chain.
func TestBlockchain_Block_FullValidationFailure(t *testing.T) {
	// Arrange
	sender := &mockBlockchainSender{}
	validator := &mockBlockValidator{
		sanityCheckResult:    true,
		validateHeaderResult: true,
		fullValidationResult: false,
		fullValidationErr:    errors.New("merkle root in header does not match calculated merkle root"),
	}
	store := &mockBlockStore{
		isOrphanResult: false,
		mainChainTip:   createTestBlock(common.Hash{}, 1),
	}
	reorg := &mockChainReorganization{}

	bc := &Blockchain{
		blockchainMsgSender: sender,
		blockValidator:      validator,
		blockStore:          store,
		chainReorganization: reorg,
		mempool:             NewMempool(nil, nil),
		observers:           mapset.NewSet[observer.BlockchainObserverAPI](),
	}

	testBlock := createTestBlock(common.Hash{}, 123)
	peerID := common.PeerId("peer-invalid")

	// Act
	bc.Block(testBlock, peerID)

	// Assert: All validations were called
	assert.True(t, validator.sanityCheckCalled, "SanityCheck should be called")
	assert.True(t, validator.validateHeaderCalled, "ValidateHeader should be called")
	assert.True(t, store.isOrphanCalled, "IsOrphanBlock should be called")
	assert.True(t, validator.fullValidationCalled, "FullValidation should be called")

	// Assert: No reorganization or broadcasting for invalid blocks
	assert.False(t, reorg.checkAndReorganizeCalled, "CheckAndReorganize should not be called when FullValidation fails")
	assert.False(t, sender.broadcastAddedBlocksCalled, "BroadcastAddedBlocks should not be called when FullValidation fails")
}

// TestBlockchain_Block_SuccessfulProcessing verifies that a valid block
// goes through all processing steps successfully.
func TestBlockchain_Block_SuccessfulProcessing(t *testing.T) {
	// Arrange
	sender := &mockBlockchainSender{}
	validator := &mockBlockValidator{
		sanityCheckResult:    true,
		validateHeaderResult: true,
		fullValidationResult: true,
	}
	store := &mockBlockStore{
		isOrphanResult: false,
		currentHeight:  5,
		mainChainTip:   createTestBlock(common.Hash{}, 1),
	}
	reorg := &mockChainReorganization{
		checkAndReorganizeResult: false,
	}

	bc := &Blockchain{
		blockchainMsgSender: sender,
		blockValidator:      validator,
		blockStore:          store,
		chainReorganization: reorg,
		mempool:             NewMempool(nil, nil),
		observers:           mapset.NewSet[observer.BlockchainObserverAPI](),
	}

	testBlock := createTestBlock(common.Hash{}, 123)
	peerID := common.PeerId("peer-valid")

	// Act
	bc.Block(testBlock, peerID)

	// Assert: All processing steps were called
	assert.True(t, validator.sanityCheckCalled, "SanityCheck should be called")
	assert.True(t, validator.validateHeaderCalled, "ValidateHeader should be called")
	assert.True(t, store.isOrphanCalled, "IsOrphanBlock should be called")
	assert.True(t, validator.fullValidationCalled, "FullValidation should be called")
	assert.True(t, store.getMainChainTipCalled, "GetMainChainTip should be called")
	assert.True(t, reorg.checkAndReorganizeCalled, "CheckAndReorganize should be called")

	// Assert: Block was added to store
	assert.Len(t, store.addedBlocks, 1, "Block should be added to store")
	assert.Equal(t, testBlock.Hash(), store.addedBlocks[0], "Added block hash should match")

	// Assert: Block was broadcast
	assert.True(t, sender.broadcastAddedBlocksCalled, "BroadcastAddedBlocks should be called")
	assert.Len(t, sender.broadcastAddedBlocksHashes, 1, "Should broadcast one block hash")
	assert.Equal(t, testBlock.Hash(), sender.broadcastAddedBlocksHashes[0], "Broadcast hash should match")
	assert.Equal(t, peerID, sender.broadcastAddedBlocksExcluded, "Excluded peer should match sender")
}

// TestBlockchain_Block_WithChainReorganization verifies that when reorganization
// occurs, the block is still broadcast correctly.
func TestBlockchain_Block_WithChainReorganization(t *testing.T) {
	// Arrange
	sender := &mockBlockchainSender{}
	validator := &mockBlockValidator{
		sanityCheckResult:    true,
		validateHeaderResult: true,
		fullValidationResult: true,
	}
	store := &mockBlockStore{
		isOrphanResult: false,
		mainChainTip:   createTestBlock(common.Hash{}, 1),
	}
	reorg := &mockChainReorganization{
		checkAndReorganizeResult: true, // Reorganization occurred
	}

	bc := &Blockchain{
		blockchainMsgSender: sender,
		blockValidator:      validator,
		blockStore:          store,
		chainReorganization: reorg,
		mempool:             NewMempool(nil, nil),
		observers:           mapset.NewSet[observer.BlockchainObserverAPI](),
	}

	testBlock := createTestBlock(common.Hash{}, 123)
	peerID := common.PeerId("peer-reorg")

	// Act
	bc.Block(testBlock, peerID)

	// Assert: Reorganization was checked
	assert.True(t, reorg.checkAndReorganizeCalled, "CheckAndReorganize should be called")

	// Assert: Block was still broadcast despite reorganization
	assert.True(t, sender.broadcastAddedBlocksCalled, "BroadcastAddedBlocks should be called even with reorganization")
}

// TestBlockchain_Block_AddedBlocksBroadcast verifies that the correct hashes
// are broadcast after block processing.
func TestBlockchain_Block_AddedBlocksBroadcast(t *testing.T) {
	// Arrange
	sender := &mockBlockchainSender{}
	validator := &mockBlockValidator{
		sanityCheckResult:    true,
		validateHeaderResult: true,
		fullValidationResult: true,
	}

	testBlock := createTestBlock(common.Hash{}, 123)
	expectedHash := testBlock.Hash()

	store := &mockBlockStore{
		isOrphanResult: false,
		mainChainTip:   createTestBlock(common.Hash{}, 1),
		// Simulate multiple blocks being added (e.g., connecting orphans)
		addBlockReturnValue: []common.Hash{expectedHash, {1, 2, 3}},
	}

	reorg := &mockChainReorganization{
		checkAndReorganizeResult: false,
	}

	bc := &Blockchain{
		blockchainMsgSender: sender,
		blockValidator:      validator,
		blockStore:          store,
		chainReorganization: reorg,
		mempool:             NewMempool(nil, nil),
		observers:           mapset.NewSet[observer.BlockchainObserverAPI](),
	}

	peerID := common.PeerId("peer-broadcast")

	// Act
	bc.Block(testBlock, peerID)

	// Assert: All added blocks were broadcast
	assert.True(t, sender.broadcastAddedBlocksCalled, "BroadcastAddedBlocks should be called")
	assert.Len(t, sender.broadcastAddedBlocksHashes, 2, "Should broadcast all added blocks")
	assert.Equal(t, expectedHash, sender.broadcastAddedBlocksHashes[0], "First broadcast hash should match")
}

// TestBlockchain_Block_ExcludedPeerInBroadcast verifies that the peer who sent
// the block is excluded from the broadcast.
func TestBlockchain_Block_ExcludedPeerInBroadcast(t *testing.T) {
	// Arrange
	sender := &mockBlockchainSender{}
	validator := &mockBlockValidator{
		sanityCheckResult:    true,
		validateHeaderResult: true,
		fullValidationResult: true,
	}
	store := &mockBlockStore{
		isOrphanResult: false,
		mainChainTip:   createTestBlock(common.Hash{}, 1),
	}
	reorg := &mockChainReorganization{
		checkAndReorganizeResult: false,
	}

	bc := &Blockchain{
		blockchainMsgSender: sender,
		blockValidator:      validator,
		blockStore:          store,
		chainReorganization: reorg,
		mempool:             NewMempool(nil, nil),
		observers:           mapset.NewSet[observer.BlockchainObserverAPI](),
	}

	testBlock := createTestBlock(common.Hash{}, 123)
	senderPeerID := common.PeerId("peer-sender")

	// Act
	bc.Block(testBlock, senderPeerID)

	// Assert: Sender peer was excluded from broadcast
	assert.True(t, sender.broadcastAddedBlocksCalled, "BroadcastAddedBlocks should be called")
	assert.Equal(t, senderPeerID, sender.broadcastAddedBlocksExcluded, "Sender peer should be excluded from broadcast")
}
