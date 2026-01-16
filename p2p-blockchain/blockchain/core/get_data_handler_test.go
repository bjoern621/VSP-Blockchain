package core

import (
	"errors"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/validation"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/inv"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockBlockchainMsgSender struct {
	sendBlockCalled bool
	lastBlock       block.Block
	lastBlockPeerID common.PeerId

	sendTxCalled bool
	lastTx       transaction.Transaction
	lastTxPeerID common.PeerId
}

func (m *mockBlockchainMsgSender) SendGetData(inventory []*inv.InvVector, peerId common.PeerId) {}

func (m *mockBlockchainMsgSender) SendInv(inventory []*inv.InvVector, peerId common.PeerId) {}

func (m *mockBlockchainMsgSender) BroadcastInvExclusionary(inventory []*inv.InvVector, peerId common.PeerId) {
}

func (m *mockBlockchainMsgSender) BroadcastAddedBlocks(blockHashes []common.Hash, excludedPeerId common.PeerId) {
}

func (m *mockBlockchainMsgSender) RequestMissingBlockHeaders(blockLocator block.BlockLocator, peerId common.PeerId) {
}

func (m *mockBlockchainMsgSender) SendHeaders(headers []*block.BlockHeader, peerId common.PeerId) {}

func (m *mockBlockchainMsgSender) SendBlock(b block.Block, peerId common.PeerId) {
	m.sendBlockCalled = true
	m.lastBlock = b
	m.lastBlockPeerID = peerId
}

func (m *mockBlockchainMsgSender) SendTx(tx transaction.Transaction, peerId common.PeerId) {
	m.sendTxCalled = true
	m.lastTx = tx
	m.lastTxPeerID = peerId
}

type mockBlockStoreGetData struct {
	blocks       map[common.Hash]block.Block
	getBlockErrs map[common.Hash]error
}

func newMockBlockStoreGetData() *mockBlockStoreGetData {
	return &mockBlockStoreGetData{
		blocks:       make(map[common.Hash]block.Block),
		getBlockErrs: make(map[common.Hash]error),
	}
}

func (m *mockBlockStoreGetData) GetBlockByHash(hash common.Hash) (block.Block, error) {
	if err, ok := m.getBlockErrs[hash]; ok {
		return block.Block{}, err
	}
	if b, ok := m.blocks[hash]; ok {
		return b, nil
	}
	return block.Block{}, errors.New("block not found")
}

func (m *mockBlockStoreGetData) AddBlock(b block.Block) []common.Hash {
	m.blocks[b.Hash()] = b
	return []common.Hash{b.Hash()}
}

func (m *mockBlockStoreGetData) IsOrphanBlock(b block.Block) (bool, error) {
	return false, nil
}

func (m *mockBlockStoreGetData) GetCurrentHeight() uint64 {
	return 0
}

func (m *mockBlockStoreGetData) GetMainChainTip() block.Block {
	return block.Block{}
}

func (m *mockBlockStoreGetData) GetBlocksByHeight(height uint64) []block.Block {
	return nil
}

func (m *mockBlockStoreGetData) IsPartOfMainChain(b block.Block) bool {
	return true
}

func createTestBlockForGetData(nonce uint32) block.Block {
	var merkleRoot common.Hash
	for i := range 32 {
		merkleRoot[i] = byte(i + 1)
	}

	return block.Block{
		Header: block.BlockHeader{
			PreviousBlockHash: [32]byte{},
			MerkleRoot:        merkleRoot,
			Timestamp:         1000 + int64(nonce),
			DifficultyTarget:  0,
			Nonce:             nonce,
		},
		Transactions: []transaction.Transaction{
			transaction.NewCoinbaseTransaction([20]byte{1, 2, 3}, 50),
		},
	}
}

func createTestTransactionForGetData() transaction.Transaction {
	return transaction.Transaction{
		Inputs: []transaction.Input{},
		Outputs: []transaction.Output{
			{
				Value:      100,
				PubKeyHash: transaction.PubKeyHash{1, 2, 3},
			},
		},
		LockTime: 0,
	}
}

func TestGetDataHandler_GetData_SendsBlock_WhenBlockRequestedAndFound(t *testing.T) {
	sender := &mockBlockchainMsgSender{}
	blockValidator := &mockBlockValidator{
		sanityCheckResult: true,
	}
	txValidator := validation.NewValidationService(nil)
	store := newMockBlockStoreGetData()
	reorg := &mockChainReorganization{}

	testBlock := createTestBlockForGetData(123)
	store.blocks[testBlock.Hash()] = testBlock

	mempool := &Mempool{
		validator:  txValidator,
		blockStore: store,
	}

	bc := &Blockchain{
		mempool:              mempool,
		blockchainMsgSender:  sender,
		transactionValidator: txValidator,
		blockValidator:       blockValidator,
		blockStore:           store,
		chainReorganization:  reorg,
	}

	inventory := []*inv.InvVector{
		{
			InvType: inv.InvTypeMsgBlock,
			Hash:    testBlock.Hash(),
		},
	}
	peerID := common.PeerId("peer-1")

	bc.GetData(inventory, peerID)

	assert.True(t, sender.sendBlockCalled, "SendBlock should be called")
	assert.Equal(t, testBlock, sender.lastBlock, "Sent block should match requested block")
	assert.Equal(t, peerID, sender.lastBlockPeerID, "PeerID should match")
	assert.False(t, sender.sendTxCalled, "SendTx should not be called")
}

func TestGetDataHandler_GetData_DoesNotSendBlock_WhenBlockRequestedButNotFound(t *testing.T) {
	sender := &mockBlockchainMsgSender{}
	blockValidator := &mockBlockValidator{
		sanityCheckResult: true,
	}
	txValidator := validation.NewValidationService(nil)
	store := newMockBlockStoreGetData()
	reorg := &mockChainReorganization{}

	mempool := &Mempool{
		validator:  txValidator,
		blockStore: store,
	}

	bc := &Blockchain{
		mempool:              mempool,
		blockchainMsgSender:  sender,
		transactionValidator: txValidator,
		blockValidator:       blockValidator,
		blockStore:           store,
		chainReorganization:  reorg,
	}

	var missingHash common.Hash
	missingHash[0] = 0xFF

	inventory := []*inv.InvVector{
		{
			InvType: inv.InvTypeMsgBlock,
			Hash:    missingHash,
		},
	}
	peerID := common.PeerId("peer-1")

	bc.GetData(inventory, peerID)

	assert.False(t, sender.sendBlockCalled, "SendBlock should not be called for missing block")
	assert.False(t, sender.sendTxCalled, "SendTx should not be called")
}

func TestGetDataHandler_GetData_SendsTransaction_WhenTransactionRequestedAndFound(t *testing.T) {
	sender := &mockBlockchainMsgSender{}
	blockValidator := &mockBlockValidator{
		sanityCheckResult: true,
	}
	txValidator := validation.NewValidationService(nil)
	store := newMockBlockStoreGetData()
	reorg := &mockChainReorganization{}

	mempool := NewMempool(txValidator, store)

	testTx := createTestTransactionForGetData()
	mempool.AddTransaction(testTx)

	bc := &Blockchain{
		mempool:              mempool,
		blockchainMsgSender:  sender,
		transactionValidator: txValidator,
		blockValidator:       blockValidator,
		blockStore:           store,
		chainReorganization:  reorg,
	}

	txId := testTx.TransactionId()
	var txHash common.Hash
	copy(txHash[:], txId[:])

	inventory := []*inv.InvVector{
		{
			InvType: inv.InvTypeMsgTx,
			Hash:    txHash,
		},
	}
	peerID := common.PeerId("peer-1")

	bc.GetData(inventory, peerID)

	assert.True(t, sender.sendTxCalled, "SendTx should be called")
	assert.Equal(t, testTx, sender.lastTx, "Sent transaction should match requested transaction")
	assert.Equal(t, peerID, sender.lastTxPeerID, "PeerID should match")
	assert.False(t, sender.sendBlockCalled, "SendBlock should not be called")
}

func TestGetDataHandler_GetData_DoesNotSendTransaction_WhenTransactionRequestedButNotFound(t *testing.T) {
	sender := &mockBlockchainMsgSender{}
	blockValidator := &mockBlockValidator{
		sanityCheckResult: true,
	}
	txValidator := validation.NewValidationService(nil)
	store := newMockBlockStoreGetData()
	reorg := &mockChainReorganization{}

	mempool := NewMempool(txValidator, store)

	bc := &Blockchain{
		mempool:              mempool,
		blockchainMsgSender:  sender,
		transactionValidator: txValidator,
		blockValidator:       blockValidator,
		blockStore:           store,
		chainReorganization:  reorg,
	}

	var missingHash common.Hash
	missingHash[0] = 0xFF

	inventory := []*inv.InvVector{
		{
			InvType: inv.InvTypeMsgTx,
			Hash:    missingHash,
		},
	}
	peerID := common.PeerId("peer-1")

	bc.GetData(inventory, peerID)

	assert.False(t, sender.sendTxCalled, "SendTx should not be called for missing transaction")
	assert.False(t, sender.sendBlockCalled, "SendBlock should not be called")
}

func TestGetDataHandler_GetData_Panics_WhenFilteredBlockRequested(t *testing.T) {
	sender := &mockBlockchainMsgSender{}
	blockValidator := &mockBlockValidator{
		sanityCheckResult: true,
	}
	txValidator := validation.NewValidationService(nil)
	store := newMockBlockStoreGetData()
	reorg := &mockChainReorganization{}

	mempool := NewMempool(txValidator, store)

	bc := &Blockchain{
		mempool:              mempool,
		blockchainMsgSender:  sender,
		transactionValidator: txValidator,
		blockValidator:       blockValidator,
		blockStore:           store,
		chainReorganization:  reorg,
	}

	var hash common.Hash
	hash[0] = 0xAB

	inventory := []*inv.InvVector{
		{
			InvType: inv.InvTypeMsgFilteredBlock,
			Hash:    hash,
		},
	}
	peerID := common.PeerId("peer-1")

	assert.Panics(t, func() {
		bc.GetData(inventory, peerID)
	}, "GetData should panic when InvTypeMsgFilteredBlock is requested")

	assert.False(t, sender.sendBlockCalled, "SendBlock should not be called")
	assert.False(t, sender.sendTxCalled, "SendTx should not be called")
}

func TestGetDataHandler_GetData_ProcessesMultipleInventoryItems(t *testing.T) {
	sender := &mockBlockchainMsgSender{}
	blockValidator := &mockBlockValidator{
		sanityCheckResult: true,
	}
	txValidator := validation.NewValidationService(nil)
	store := newMockBlockStoreGetData()
	reorg := &mockChainReorganization{}

	testBlock1 := createTestBlockForGetData(1)
	testBlock2 := createTestBlockForGetData(2)
	store.blocks[testBlock1.Hash()] = testBlock1
	store.blocks[testBlock2.Hash()] = testBlock2

	mempool := NewMempool(txValidator, store)

	testTx := createTestTransactionForGetData()
	mempool.AddTransaction(testTx)

	bc := &Blockchain{
		mempool:              mempool,
		blockchainMsgSender:  sender,
		transactionValidator: txValidator,
		blockValidator:       blockValidator,
		blockStore:           store,
		chainReorganization:  reorg,
	}

	txId := testTx.TransactionId()
	var txHash common.Hash
	copy(txHash[:], txId[:])

	inventory := []*inv.InvVector{
		{
			InvType: inv.InvTypeMsgBlock,
			Hash:    testBlock1.Hash(),
		},
		{
			InvType: inv.InvTypeMsgTx,
			Hash:    txHash,
		},
		{
			InvType: inv.InvTypeMsgBlock,
			Hash:    testBlock2.Hash(),
		},
	}
	peerID := common.PeerId("peer-1")

	bc.GetData(inventory, peerID)

	assert.True(t, sender.sendBlockCalled, "SendBlock should be called")
	assert.True(t, sender.sendTxCalled, "SendTx should be called")
	assert.Equal(t, peerID, sender.lastBlockPeerID, "Last block peer ID should match")
	assert.Equal(t, peerID, sender.lastTxPeerID, "Last tx peer ID should match")
}

func TestGetDataHandler_GetData_HandlesMixedFoundAndNotFoundItems(t *testing.T) {
	sender := &mockBlockchainMsgSender{}
	blockValidator := &mockBlockValidator{
		sanityCheckResult: true,
	}
	txValidator := validation.NewValidationService(nil)
	store := newMockBlockStoreGetData()
	reorg := &mockChainReorganization{}

	testBlock := createTestBlockForGetData(123)
	store.blocks[testBlock.Hash()] = testBlock

	mempool := NewMempool(txValidator, store)

	bc := &Blockchain{
		mempool:              mempool,
		blockchainMsgSender:  sender,
		transactionValidator: txValidator,
		blockValidator:       blockValidator,
		blockStore:           store,
		chainReorganization:  reorg,
	}

	var missingTxHash common.Hash
	missingTxHash[0] = 0xFF

	inventory := []*inv.InvVector{
		{
			InvType: inv.InvTypeMsgBlock,
			Hash:    testBlock.Hash(),
		},
		{
			InvType: inv.InvTypeMsgTx,
			Hash:    missingTxHash,
		},
	}
	peerID := common.PeerId("peer-1")

	bc.GetData(inventory, peerID)

	assert.True(t, sender.sendBlockCalled, "SendBlock should be called for found block")
	assert.False(t, sender.sendTxCalled, "SendTx should not be called for missing transaction")
}

func TestGetDataHandler_GetData_HandlesEmptyInventory(t *testing.T) {
	sender := &mockBlockchainMsgSender{}
	blockValidator := &mockBlockValidator{
		sanityCheckResult: true,
	}
	txValidator := validation.NewValidationService(nil)
	store := newMockBlockStoreGetData()
	reorg := &mockChainReorganization{}

	mempool := NewMempool(txValidator, store)

	bc := &Blockchain{
		mempool:              mempool,
		blockchainMsgSender:  sender,
		transactionValidator: txValidator,
		blockValidator:       blockValidator,
		blockStore:           store,
		chainReorganization:  reorg,
	}

	inventory := []*inv.InvVector{}
	peerID := common.PeerId("peer-1")

	bc.GetData(inventory, peerID)

	assert.False(t, sender.sendBlockCalled, "SendBlock should not be called")
	assert.False(t, sender.sendTxCalled, "SendTx should not be called")
}
