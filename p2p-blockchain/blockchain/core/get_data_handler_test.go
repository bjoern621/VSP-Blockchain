package core

import (
	"errors"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/validation"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/inv"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockUtxoStoreForGetData is a mock UTXO store for GetData handler tests
type mockUtxoStoreForGetData struct {
	utxos map[string]transaction.Output // key: txID+outputIndex
}

func newMockUtxoStoreForGetData() *mockUtxoStoreForGetData {
	return &mockUtxoStoreForGetData{
		utxos: make(map[string]transaction.Output),
	}
}

func (m *mockUtxoStoreForGetData) AddUtxo(txID transaction.TransactionID, outputIndex uint32, output transaction.Output) {
	key := string(txID[:]) + string(rune(outputIndex))
	m.utxos[key] = output
}

func (m *mockUtxoStoreForGetData) InitializeGenesisPool(_ block.Block) error {
	return nil
}

func (m *mockUtxoStoreForGetData) AddNewBlock(_ block.Block) error {
	return nil
}

func (m *mockUtxoStoreForGetData) GetUtxoFromBlock(txID transaction.TransactionID, outputIndex uint32, _ common.Hash) (transaction.Output, error) {
	key := string(txID[:]) + string(rune(outputIndex))
	if output, ok := m.utxos[key]; ok {
		return output, nil
	}
	return transaction.Output{}, errors.New("UTXO not found")
}

func (m *mockUtxoStoreForGetData) ValidateBlock(_ block.Block) bool {
	return true
}

func (m *mockUtxoStoreForGetData) ValidateTransactionFromBlock(_ transaction.Transaction, _ common.Hash) bool {
	return true
}

func (m *mockUtxoStoreForGetData) GetUtxosByPubKeyHashFromBlock(_ transaction.PubKeyHash, _ common.Hash) ([]transaction.UTXO, error) {
	return []transaction.UTXO{}, nil
}

type mockBlockchainMsgSender struct {
	mu              sync.RWMutex
	sendBlockCalled bool
	lastBlock       block.Block
	lastBlockPeerID common.PeerId

	sendTxCalled bool
	lastTx       transaction.Transaction
	lastTxPeerID common.PeerId

	blockDone chan int
	txDone    chan int
}

func (m *mockBlockchainMsgSender) SendBlock(b block.Block, peerId common.PeerId) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sendBlockCalled = true
	m.lastBlock = b
	m.lastBlockPeerID = peerId
	if m.blockDone != nil {
		m.blockDone <- 1
	}
}

func (m *mockBlockchainMsgSender) SendTx(tx transaction.Transaction, peerId common.PeerId) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sendTxCalled = true
	m.lastTx = tx
	m.lastTxPeerID = peerId
	if m.txDone != nil {
		m.txDone <- 1
	}
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

func (m *mockBlockStoreGetData) IsOrphanBlock(_ block.Block) (bool, error) {
	return false, nil
}

func (m *mockBlockStoreGetData) GetCurrentHeight() uint64 {
	return 0
}

func (m *mockBlockStoreGetData) GetMainChainTip() block.Block {
	return block.Block{}
}

func (m *mockBlockStoreGetData) GetBlocksByHeight(_ uint64) []block.Block {
	return nil
}

func (m *mockBlockStoreGetData) IsPartOfMainChain(_ block.Block) bool {
	return true
}

func (m *mockBlockStoreGetData) GetAllBlocksWithMetadata() []block.BlockWithMetadata {
	return nil
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
			transaction.NewCoinbaseTransaction([20]byte{1, 2, 3}, 50, 1),
		},
	}
}

func createTestTransactionForGetData() transaction.Transaction {
	// Create a PubKey for the transaction input
	pubKey := transaction.PubKey{0x03, 0x04}

	// The PubKeyHash should be HASH160(pubKey) for validation to pass
	// For testing, we'll compute it properly
	pubKeyHash := transaction.Hash160(pubKey)

	return transaction.Transaction{
		Inputs: []transaction.Input{
			{
				PrevTxID:    transaction.TransactionID{1, 2, 3, 4, 5},
				OutputIndex: 0,
				Signature:   []byte{0x01, 0x02},
				PubKey:      pubKey,
			},
		},
		Outputs: []transaction.Output{
			{
				Value:      50, // Less than the input value (100) to account for fee
				PubKeyHash: pubKeyHash,
			},
		},
	}
}

// Helper function to get the expected UTXO for the test transaction
func getTestUtxoForTransaction() (transaction.TransactionID, uint32, transaction.Output) {
	pubKey := transaction.PubKey{0x03, 0x04}
	pubKeyHash := transaction.Hash160(pubKey)

	txID := transaction.TransactionID{1, 2, 3, 4, 5}
	outputIndex := uint32(0)
	output := transaction.Output{
		Value:      100,
		PubKeyHash: pubKeyHash,
	}

	return txID, outputIndex, output
}

func TestGetDataHandler_GetData_SendsBlock_WhenBlockRequestedAndFound(t *testing.T) {
	sender := &mockBlockchainMsgSender{
		blockDone: make(chan int, 1),
	}
	blockValidator := &mockBlockValidator{
		sanityCheckResult: true,
	}
	txValidator := validation.NewTransactionValidator(nil)
	store := newMockBlockStoreGetData()
	reorg := &mockChainReorganization{}

	testBlock := createTestBlockForGetData(123)
	store.blocks[testBlock.Hash()] = testBlock

	mempool := &Mempool{
		validator:  txValidator,
		blockStore: store,
	}

	bc := &Blockchain{
		mempool:                mempool,
		fullInventoryMsgSender: sender,
		transactionValidator:   txValidator,
		blockValidator:         blockValidator,
		blockStore:             store,
		chainReorganization:    reorg,
	}

	inventory := []*inv.InvVector{
		{
			InvType: inv.InvTypeMsgBlock,
			Hash:    testBlock.Hash(),
		},
	}
	peerID := common.PeerId("peer-1")

	bc.GetData(inventory, peerID)
	<-sender.blockDone

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
	txValidator := validation.NewTransactionValidator(nil)
	store := newMockBlockStoreGetData()
	reorg := &mockChainReorganization{}

	mempool := &Mempool{
		validator:  txValidator,
		blockStore: store,
	}

	bc := &Blockchain{
		mempool:                mempool,
		fullInventoryMsgSender: sender,
		transactionValidator:   txValidator,
		blockValidator:         blockValidator,
		blockStore:             store,
		chainReorganization:    reorg,
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
	sender := &mockBlockchainMsgSender{
		txDone: make(chan int, 1),
	}
	blockValidator := &mockBlockValidator{
		sanityCheckResult: true,
	}

	// Create mock UTXO store and set up the UTXO that the test transaction references
	utxoStore := newMockUtxoStoreForGetData()
	txID, outputIndex, output := getTestUtxoForTransaction()
	utxoStore.AddUtxo(txID, outputIndex, output)

	txValidator := validation.NewTransactionValidator(utxoStore)
	store := newMockBlockStoreGetData()
	reorg := &mockChainReorganization{}

	mempool := NewMempool(txValidator, store)

	testTx := createTestTransactionForGetData()
	mempool.AddTransaction(testTx)

	bc := &Blockchain{
		mempool:                mempool,
		fullInventoryMsgSender: sender,
		transactionValidator:   txValidator,
		blockValidator:         blockValidator,
		blockStore:             store,
		chainReorganization:    reorg,
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
	<-sender.txDone

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
	txValidator := validation.NewTransactionValidator(nil)
	store := newMockBlockStoreGetData()
	reorg := &mockChainReorganization{}

	mempool := NewMempool(txValidator, store)

	bc := &Blockchain{
		mempool:                mempool,
		fullInventoryMsgSender: sender,
		transactionValidator:   txValidator,
		blockValidator:         blockValidator,
		blockStore:             store,
		chainReorganization:    reorg,
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
	txValidator := validation.NewTransactionValidator(nil)
	store := newMockBlockStoreGetData()
	reorg := &mockChainReorganization{}

	mempool := NewMempool(txValidator, store)

	bc := &Blockchain{
		mempool:                mempool,
		fullInventoryMsgSender: sender,
		transactionValidator:   txValidator,
		blockValidator:         blockValidator,
		blockStore:             store,
		chainReorganization:    reorg,
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
	sender := &mockBlockchainMsgSender{
		blockDone: make(chan int, 2),
		txDone:    make(chan int, 1),
	}
	blockValidator := &mockBlockValidator{
		sanityCheckResult: true,
	}

	// Create mock UTXO store and set up the UTXO that the test transaction references
	utxoStore := newMockUtxoStoreForGetData()
	txID, outputIndex, output := getTestUtxoForTransaction()
	utxoStore.AddUtxo(txID, outputIndex, output)

	txValidator := validation.NewTransactionValidator(utxoStore)
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
		mempool:                mempool,
		fullInventoryMsgSender: sender,
		transactionValidator:   txValidator,
		blockValidator:         blockValidator,
		blockStore:             store,
		chainReorganization:    reorg,
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
	<-sender.txDone
	<-sender.blockDone
	<-sender.blockDone

	assert.True(t, sender.sendBlockCalled, "SendBlock should be called")
	assert.True(t, sender.sendTxCalled, "SendTx should be called")
	assert.Equal(t, peerID, sender.lastBlockPeerID, "Last block peer ID should match")
	assert.Equal(t, peerID, sender.lastTxPeerID, "Last tx peer ID should match")
}

func TestGetDataHandler_GetData_HandlesMixedFoundAndNotFoundItems(t *testing.T) {
	sender := &mockBlockchainMsgSender{
		blockDone: make(chan int, 1),
	}
	blockValidator := &mockBlockValidator{
		sanityCheckResult: true,
	}
	txValidator := validation.NewTransactionValidator(nil)
	store := newMockBlockStoreGetData()
	reorg := &mockChainReorganization{}

	testBlock := createTestBlockForGetData(123)
	store.blocks[testBlock.Hash()] = testBlock

	mempool := NewMempool(txValidator, store)

	bc := &Blockchain{
		mempool:                mempool,
		fullInventoryMsgSender: sender,
		transactionValidator:   txValidator,
		blockValidator:         blockValidator,
		blockStore:             store,
		chainReorganization:    reorg,
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
	<-sender.blockDone

	assert.True(t, sender.sendBlockCalled, "SendBlock should be called for found block")
	assert.False(t, sender.sendTxCalled, "SendTx should not be called for missing transaction")
}

func TestGetDataHandler_GetData_HandlesEmptyInventory(t *testing.T) {
	sender := &mockBlockchainMsgSender{}
	blockValidator := &mockBlockValidator{
		sanityCheckResult: true,
	}
	txValidator := validation.NewTransactionValidator(nil)
	store := newMockBlockStoreGetData()
	reorg := &mockChainReorganization{}

	mempool := NewMempool(txValidator, store)

	bc := &Blockchain{
		mempool:                mempool,
		fullInventoryMsgSender: sender,
		transactionValidator:   txValidator,
		blockValidator:         blockValidator,
		blockStore:             store,
		chainReorganization:    reorg,
	}

	inventory := []*inv.InvVector{}
	peerID := common.PeerId("peer-1")

	bc.GetData(inventory, peerID)

	assert.False(t, sender.sendBlockCalled, "SendBlock should not be called")
	assert.False(t, sender.sendTxCalled, "SendTx should not be called")
}
