package utxo

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"testing"

	"github.com/stretchr/testify/assert"
)

// =========================================================================
// Mock BlockStore
// =========================================================================

type mockBlockStore struct {
	orphanBlocks map[common.Hash]bool
}

func newMockBlockStore() *mockBlockStore {
	return &mockBlockStore{
		orphanBlocks: make(map[common.Hash]bool),
	}
}

func (m *mockBlockStore) IsOrphanBlock(b block.Block) (bool, error) {
	return m.orphanBlocks[b.Hash()], nil
}

func (m *mockBlockStore) AddBlock(_ block.Block) []common.Hash {
	return nil
}

func (m *mockBlockStore) IsPartOfMainChain(_ block.Block) bool {
	return true
}

func (m *mockBlockStore) GetBlockByHash(_ common.Hash) (block.Block, error) {
	return block.Block{}, nil
}

func (m *mockBlockStore) GetBlocksByHeight(_ uint64) []block.Block {
	return nil
}

func (m *mockBlockStore) GetCurrentHeight() uint64 {
	return 0
}

func (m *mockBlockStore) GetMainChainTip() block.Block {
	return block.Block{}
}

func (m *mockBlockStore) GetAllBlocksWithMetadata() []block.BlockWithMetadata {
	return nil
}

func (m *mockBlockStore) IsBlockInvalid(_ block.Block) (bool, error) {
	return false, nil
}

// =========================================================================
// Test Helpers
// =========================================================================

func createTestBlock(prevHash common.Hash, txs []transaction.Transaction) block.Block {
	header := block.BlockHeader{
		PreviousBlockHash: prevHash,
		Timestamp:         1000,
		DifficultyTarget:  10,
		Nonce:             1,
	}
	return block.Block{
		Header:       header,
		Transactions: txs,
	}
}

func createCoinbaseTx(pubKeyHash transaction.PubKeyHash, value uint64) transaction.Transaction {
	return transaction.NewCoinbaseTransaction(pubKeyHash, value, 1)
}

func createRegularTx(prevTxID transaction.TransactionID, outputIndex uint32, value uint64, pubKeyHash transaction.PubKeyHash) transaction.Transaction {
	return transaction.Transaction{
		Inputs: []transaction.Input{
			{
				PrevTxID:    prevTxID,
				OutputIndex: outputIndex,
				Signature:   []byte("sig"),
				PubKey:      transaction.PubKey{1, 2, 3},
			},
		},
		Outputs: []transaction.Output{
			{
				Value:      value,
				PubKeyHash: pubKeyHash,
			},
		},
	}
}

// =========================================================================
// Tests
// =========================================================================

func TestNewUtxoStore(t *testing.T) {
	mockStore := newMockBlockStore()
	utxoStore := NewUtxoStore(mockStore)

	assert.NotNil(t, utxoStore)
}

func TestAddNewBlock_SkipsOrphanBlock(t *testing.T) {
	// Arrange
	mockStore := newMockBlockStore()
	utxoStore := NewUtxoStore(mockStore).(*UtxoStore)

	coinbaseTx := createCoinbaseTx(transaction.PubKeyHash{1, 2, 3}, 50)
	orphanBlock := createTestBlock(common.Hash{0xff}, []transaction.Transaction{coinbaseTx})

	// Mark block as orphan
	mockStore.orphanBlocks[orphanBlock.Hash()] = true

	// Act
	err := utxoStore.AddNewBlock(orphanBlock)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, utxoStore.blockHashToPool)
}

func TestAddNewBlock_SkipsDuplicateBlock(t *testing.T) {
	// Arrange
	mockStore := newMockBlockStore()
	utxoStore := NewUtxoStore(mockStore).(*UtxoStore)

	genesisHash := common.Hash{}
	coinbaseTx := createCoinbaseTx(transaction.PubKeyHash{1, 2, 3}, 50)
	testBlock := createTestBlock(genesisHash, []transaction.Transaction{coinbaseTx})

	// Pre-populate with genesis pool and the block's pool
	utxoStore.blockHashToPool[genesisHash] = utxoPool{UtxoData: make(map[outpoint]transaction.Output)}
	utxoStore.blockHashToPool[testBlock.Hash()] = utxoPool{UtxoData: make(map[outpoint]transaction.Output)}

	initialPoolCount := len(utxoStore.blockHashToPool)

	// Act
	err := utxoStore.AddNewBlock(testBlock)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, initialPoolCount, len(utxoStore.blockHashToPool))
}

func TestAddNewBlock_AddsCoinbaseOutputsToPool(t *testing.T) {
	// Arrange
	mockStore := newMockBlockStore()
	utxoStore := NewUtxoStore(mockStore).(*UtxoStore)

	genesisHash := common.Hash{}
	pubKeyHash := transaction.PubKeyHash{1, 2, 3}
	coinbaseTx := createCoinbaseTx(pubKeyHash, 50)
	testBlock := createTestBlock(genesisHash, []transaction.Transaction{coinbaseTx})

	// Pre-populate with empty genesis pool
	utxoStore.blockHashToPool[genesisHash] = utxoPool{UtxoData: make(map[outpoint]transaction.Output)}

	// Act
	err := utxoStore.AddNewBlock(testBlock)

	// Assert
	assert.NoError(t, err)

	newPool := utxoStore.blockHashToPool[testBlock.Hash()]
	assert.Len(t, newPool.UtxoData, 1)

	// Check the coinbase output is in the pool
	txID := coinbaseTx.TransactionId()
	outpoint := outpoint{TxID: txID, OutputIndex: 0}
	output, exists := newPool.UtxoData[outpoint]
	assert.True(t, exists)
	assert.Equal(t, uint64(50), output.Value)
	assert.Equal(t, pubKeyHash, output.PubKeyHash)
}

func TestAddNewBlock_RemovesSpentUtxos(t *testing.T) {
	// Arrange
	mockStore := newMockBlockStore()
	utxoStore := NewUtxoStore(mockStore).(*UtxoStore)

	genesisHash := common.Hash{}
	pubKeyHash := transaction.PubKeyHash{1, 2, 3}

	// Create a previous UTXO
	prevTxID := transaction.TransactionID{0x01, 0x02, 0x03}
	prevoutpoint := outpoint{TxID: prevTxID, OutputIndex: 0}
	prevOutput := transaction.Output{Value: 100, PubKeyHash: pubKeyHash}

	// Pre-populate genesis pool with existing UTXO
	genesisPool := utxoPool{
		UtxoData: map[outpoint]transaction.Output{
			prevoutpoint: prevOutput,
		},
	}
	utxoStore.blockHashToPool[genesisHash] = genesisPool

	// Create a block that spends the previous UTXO
	spendingTx := createRegularTx(prevTxID, 0, 90, pubKeyHash)
	testBlock := createTestBlock(genesisHash, []transaction.Transaction{spendingTx})

	// Act
	err := utxoStore.AddNewBlock(testBlock)

	// Assert
	assert.NoError(t, err)

	newPool := utxoStore.blockHashToPool[testBlock.Hash()]

	// Previous UTXO should be removed
	_, exists := newPool.UtxoData[prevoutpoint]
	assert.False(t, exists, "spent UTXO should be removed")

	// New output should be added
	newTxID := spendingTx.TransactionId()
	newoutpoint := outpoint{TxID: newTxID, OutputIndex: 0}
	newOutput, exists := newPool.UtxoData[newoutpoint]
	assert.True(t, exists, "new UTXO should be added")
	assert.Equal(t, uint64(90), newOutput.Value)
}

func TestAddNewBlock_HandlesMultipleOutputs(t *testing.T) {
	// Arrange
	mockStore := newMockBlockStore()
	utxoStore := NewUtxoStore(mockStore).(*UtxoStore)

	genesisHash := common.Hash{}
	pubKeyHash1 := transaction.PubKeyHash{1, 2, 3}
	pubKeyHash2 := transaction.PubKeyHash{4, 5, 6}

	// Create a previous UTXO
	prevTxID := transaction.TransactionID{0x01, 0x02, 0x03}
	prevoutpoint := outpoint{TxID: prevTxID, OutputIndex: 0}
	prevOutput := transaction.Output{Value: 100, PubKeyHash: pubKeyHash1}

	genesisPool := utxoPool{
		UtxoData: map[outpoint]transaction.Output{
			prevoutpoint: prevOutput,
		},
	}
	utxoStore.blockHashToPool[genesisHash] = genesisPool

	// Create a transaction with multiple outputs
	tx := transaction.Transaction{
		Inputs: []transaction.Input{
			{
				PrevTxID:    prevTxID,
				OutputIndex: 0,
				Signature:   []byte("sig"),
				PubKey:      transaction.PubKey{1, 2, 3},
			},
		},
		Outputs: []transaction.Output{
			{Value: 60, PubKeyHash: pubKeyHash1},
			{Value: 30, PubKeyHash: pubKeyHash2},
		},
	}
	testBlock := createTestBlock(genesisHash, []transaction.Transaction{tx})

	// Act
	err := utxoStore.AddNewBlock(testBlock)

	// Assert
	assert.NoError(t, err)

	newPool := utxoStore.blockHashToPool[testBlock.Hash()]
	assert.Len(t, newPool.UtxoData, 2)

	txID := tx.TransactionId()
	output0, exists := newPool.UtxoData[outpoint{TxID: txID, OutputIndex: 0}]
	assert.True(t, exists)
	assert.Equal(t, uint64(60), output0.Value)

	output1, exists := newPool.UtxoData[outpoint{TxID: txID, OutputIndex: 1}]
	assert.True(t, exists)
	assert.Equal(t, uint64(30), output1.Value)
}

func TestAddNewBlock_PreservesPreviousPool(t *testing.T) {
	// Arrange
	mockStore := newMockBlockStore()
	utxoStore := NewUtxoStore(mockStore).(*UtxoStore)

	genesisHash := common.Hash{}
	pubKeyHash := transaction.PubKeyHash{1, 2, 3}

	// Create existing UTXOs in genesis pool
	existingTxID := transaction.TransactionID{0xaa, 0xbb, 0xcc}
	existingoutpoint := outpoint{TxID: existingTxID, OutputIndex: 0}
	existingOutput := transaction.Output{Value: 200, PubKeyHash: pubKeyHash}

	genesisPool := utxoPool{
		UtxoData: map[outpoint]transaction.Output{
			existingoutpoint: existingOutput,
		},
	}
	utxoStore.blockHashToPool[genesisHash] = genesisPool

	// Create a block with only coinbase (doesn't spend existing UTXO)
	coinbaseTx := createCoinbaseTx(pubKeyHash, 50)
	testBlock := createTestBlock(genesisHash, []transaction.Transaction{coinbaseTx})

	// Act
	err := utxoStore.AddNewBlock(testBlock)

	// Assert
	assert.NoError(t, err)

	newPool := utxoStore.blockHashToPool[testBlock.Hash()]

	// Previous UTXO should still exist
	_, exists := newPool.UtxoData[existingoutpoint]
	assert.True(t, exists, "unspent UTXO should be preserved")

	// Plus the new coinbase output
	assert.Len(t, newPool.UtxoData, 2)
}

func TestCreateUtxoPoolFromBlock_EmptyPrevPool(t *testing.T) {
	// Arrange
	mockStore := newMockBlockStore()
	utxoStore := NewUtxoStore(mockStore).(*UtxoStore)

	prevPool := utxoPool{UtxoData: make(map[outpoint]transaction.Output)}
	pubKeyHash := transaction.PubKeyHash{1, 2, 3}

	coinbaseTx := createCoinbaseTx(pubKeyHash, 50)
	testBlock := createTestBlock(common.Hash{}, []transaction.Transaction{coinbaseTx})

	// Act
	newPool := utxoStore.createUtxoPoolFromBlock(&prevPool, testBlock)

	// Assert
	assert.Len(t, newPool.UtxoData, 1)

	txID := coinbaseTx.TransactionId()
	outpoint := outpoint{TxID: txID, OutputIndex: 0}
	output, exists := newPool.UtxoData[outpoint]
	assert.True(t, exists)
	assert.Equal(t, uint64(50), output.Value)
}

func TestOutpoint_equality(t *testing.T) {
	txID1 := transaction.TransactionID{0x01, 0x02, 0x03}
	txID2 := transaction.TransactionID{0x01, 0x02, 0x03}
	txID3 := transaction.TransactionID{0x04, 0x05, 0x06}

	outpoint1 := outpoint{TxID: txID1, OutputIndex: 0}
	outpoint2 := outpoint{TxID: txID2, OutputIndex: 0}
	outpoint3 := outpoint{TxID: txID1, OutputIndex: 1}
	outpoint4 := outpoint{TxID: txID3, OutputIndex: 0}

	assert.Equal(t, outpoint1, outpoint2, "same TxID and OutputIndex should be equal")
	assert.NotEqual(t, outpoint1, outpoint3, "different OutputIndex should not be equal")
	assert.NotEqual(t, outpoint1, outpoint4, "different TxID should not be equal")
}
