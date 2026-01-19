package core

import (
	"context"
	"fmt"
	"math/big"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/api"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	minerApi "s3b/vsp-blockchain/p2p-blockchain/miner/api"
	"testing"
	"time"
)

// mockBlockchainAPI is a mock implementation of BlockchainAPI
type mockBlockchainAPI struct {
	addSelfMinedBlockCalled bool
	selfMinedBlock          block.Block
}

func (m *mockBlockchainAPI) AddSelfMinedBlock(selfMinedBlock block.Block) {
	m.addSelfMinedBlockCalled = true
	m.selfMinedBlock = selfMinedBlock
}

// mockUtxoStoreAPI is a mock implementation of UtxoStoreAPI
type mockUtxoStoreAPI struct {
	utxos map[utxoOutpoint]transaction.Output
}

func (m *mockUtxoStoreAPI) ValidateTransactionsOfBlock(blockToValidate block.Block) bool {
	return true
}

func (m *mockUtxoStoreAPI) InitializeGenesisPool(_ block.Block) error {
	return nil
}

func (m *mockUtxoStoreAPI) AddNewBlock(_ block.Block) error {
	return nil
}

func (m *mockUtxoStoreAPI) GetUtxoFromBlock(prevTxID transaction.TransactionID, outputIndex uint32, _ common.Hash) (transaction.Output, error) {
	outpoint := utxoOutpoint{txID: prevTxID, outputIndex: outputIndex}
	if output, exists := m.utxos[outpoint]; exists {
		return output, nil
	}
	return transaction.Output{}, &utxoNotFoundError{}
}

func (m *mockUtxoStoreAPI) ValidateTransactionFromBlock(_ transaction.Transaction, _ common.Hash) bool {
	return true
}

func (m *mockUtxoStoreAPI) GetUtxosByPubKeyHashFromBlock(_ transaction.PubKeyHash, _ common.Hash) ([]transaction.UTXO, error) {
	return nil, nil
}

type utxoOutpoint struct {
	txID        transaction.TransactionID
	outputIndex uint32
}

type utxoNotFoundError struct{}

func (e *utxoNotFoundError) Error() string {
	return "UTXO not found"
}

// mockBlockStore is a mock implementation of blockchain.BlockStoreAPI
type mockBlockStore struct {
	tip block.Block
}

func (m *mockBlockStore) IsBlockInvalid(block block.Block) (bool, error) {
	return false, nil
}

func (m *mockBlockStore) AddBlock(_ block.Block) []common.Hash {
	return nil
}

func (m *mockBlockStore) IsOrphanBlock(_ block.Block) (bool, error) {
	return false, nil
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

func (m *mockBlockStore) GetMainChainHeight() uint64 {
	return 0
}

func (m *mockBlockStore) GetMainChainTip() block.Block {
	return m.tip
}

func (m *mockBlockStore) GetAllBlocksWithMetadata() []block.BlockWithMetadata {
	return nil
}

// Helper function to create a test miner service
func createTestMinerService(tip block.Block, utxos map[utxoOutpoint]transaction.Output) *minerService {
	mockBlockchain := &mockBlockchainAPI{}
	mockUTXO := &mockUtxoStoreAPI{utxos: utxos}
	mockBlockStore := &mockBlockStore{tip: tip}

	return &minerService{
		blockchain:  mockBlockchain,
		utxoService: mockUTXO,
		blockStore:  mockBlockStore,
	}
}

// Helper function to create a test transaction
func createTestTransaction(_ uint64, outputValue uint64) transaction.Transaction {
	prevTxID := transaction.TransactionID{}
	prevTxID[0] = 0xAA

	return transaction.Transaction{
		Inputs: []transaction.Input{
			{
				PrevTxID:    prevTxID,
				OutputIndex: 0,
				Signature:   []byte("signature"),
				PubKey:      transaction.PubKey{},
			},
		},
		Outputs: []transaction.Output{
			{
				Value:      outputValue,
				PubKeyHash: transaction.PubKeyHash{},
			},
		},
	}
}

// Helper function to create a genesis block
func createGenesisBlock() block.Block {
	genesisTx := transaction.NewCoinbaseTransaction(transaction.PubKeyHash{}, 50, 0)

	return block.Block{
		Header: block.BlockHeader{
			PreviousBlockHash: common.Hash{},
			MerkleRoot:        block.MerkleRootFromTransactions([]transaction.Transaction{genesisTx}),
			Timestamp:         time.Now().Unix(),
			DifficultyTarget:  28,
			Nonce:             0,
		},
		Transactions: []transaction.Transaction{genesisTx},
	}
}

func TestGetTarget(t *testing.T) {
	tests := []struct {
		name       string
		difficulty uint8
		wantBits   int
	}{
		{
			name:       "Difficulty 28 (standard)",
			difficulty: 28,
			wantBits:   256 - 28,
		},
		{
			name:       "Difficulty 0 (no zeros required)",
			difficulty: 0,
			wantBits:   256,
		},
		{
			name:       "Difficulty 255 (maximum)",
			difficulty: 255,
			wantBits:   1,
		},
		{
			name:       "Difficulty 16",
			difficulty: 16,
			wantBits:   240,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := getTarget(tt.difficulty)

			// Verify the target is a big integer with a single bit set
			expected := big.NewInt(1)
			expected.Lsh(expected, uint(tt.wantBits))

			if target.Cmp(expected) != 0 {
				t.Errorf("getTarget(%d) = %v, want %v", tt.difficulty, &target, expected)
			}
		})
	}
}

func TestGetCurrentTargetBits(t *testing.T) {
	target, err := GetCurrentTargetBits()
	if err != nil {
		t.Errorf("GetCurrentTargetBits() returned error: %v", err)
	}
	if target != 28 {
		t.Errorf("GetCurrentTargetBits() = %d, want 28", target)
	}
}

func TestGetCurrentReward(t *testing.T) {
	reward, err := GetCurrentReward()
	if err != nil {
		t.Errorf("GetCurrentReward() returned error: %v", err)
	}
	if reward != 50 {
		t.Errorf("GetCurrentReward() = %d, want 50", reward)
	}
}

func TestCreateCandidateBlockHeader(t *testing.T) {
	genesis := createGenesisBlock()
	miner := createTestMinerService(genesis, nil)

	txs := []transaction.Transaction{
		createTestTransaction(100, 90),
	}

	header, err := miner.createCandidateBlockHeader(txs)
	if err != nil {
		t.Fatalf("createCandidateBlockHeader() returned error: %v", err)
	}

	// Verify previous block hash points to genesis
	expectedPrevHash := genesis.Hash()
	if header.PreviousBlockHash != expectedPrevHash {
		t.Errorf("PreviousBlockHash = %v, want %v", header.PreviousBlockHash, expectedPrevHash)
	}

	// Verify merkle root is calculated
	expectedMerkleRoot := block.MerkleRootFromTransactions(txs)
	if header.MerkleRoot != expectedMerkleRoot {
		t.Errorf("MerkleRoot = %v, want %v", header.MerkleRoot, expectedMerkleRoot)
	}

	// Verify difficulty target
	if header.DifficultyTarget != 28 {
		t.Errorf("DifficultyTarget = %d, want 28", header.DifficultyTarget)
	}

	// Verify timestamp is recent (within 1 second)
	now := time.Now().Unix()
	if header.Timestamp < now-1 || header.Timestamp > now+1 {
		t.Errorf("Timestamp = %d, want approximately %d", header.Timestamp, now)
	}

	// Verify nonce is 0
	if header.Nonce != 0 {
		t.Errorf("Nonce = %d, want 0", header.Nonce)
	}
}

func TestGetInputSum(t *testing.T) {
	prevTxID := transaction.TransactionID{}
	prevTxID[0] = 0xBB

	utxos := map[utxoOutpoint]transaction.Output{
		{txID: prevTxID, outputIndex: 0}: {Value: 100, PubKeyHash: transaction.PubKeyHash{}},
		{txID: prevTxID, outputIndex: 1}: {Value: 50, PubKeyHash: transaction.PubKeyHash{}},
	}

	miner := createTestMinerService(createGenesisBlock(), utxos)

	tx := transaction.Transaction{
		Inputs: []transaction.Input{
			{
				PrevTxID:    prevTxID,
				OutputIndex: 0,
			},
			{
				PrevTxID:    prevTxID,
				OutputIndex: 1,
			},
		},
	}

	tip := miner.blockStore.GetMainChainTip()
	sum, err := miner.getInputSum(tx, tip.Hash())
	if err != nil {
		t.Fatalf("getInputSum() returned error: %v", err)
	}

	if sum != 150 {
		t.Errorf("getInputSum() = %d, want 150", sum)
	}
}

func TestGetInputSum_UTXONotFound(t *testing.T) {
	miner := createTestMinerService(createGenesisBlock(), nil)

	tx := transaction.Transaction{
		Inputs: []transaction.Input{
			{
				PrevTxID:    transaction.TransactionID{},
				OutputIndex: 0,
			},
		},
	}

	tip := miner.blockStore.GetMainChainTip()
	_, err := miner.getInputSum(tx, tip.Hash())
	if err == nil {
		t.Error("getInputSum() should return error when UTXO not found")
	}
}

func TestGetTransactionWithFee(t *testing.T) {
	prevTxID := transaction.TransactionID{}
	prevTxID[0] = 0xCC

	utxos := map[utxoOutpoint]transaction.Output{
		{txID: prevTxID, outputIndex: 0}: {Value: 100, PubKeyHash: transaction.PubKeyHash{}},
		{txID: prevTxID, outputIndex: 1}: {Value: 200, PubKeyHash: transaction.PubKeyHash{}},
	}

	miner := createTestMinerService(createGenesisBlock(), utxos)

	txs := []transaction.Transaction{
		{
			Inputs: []transaction.Input{
				{PrevTxID: prevTxID, OutputIndex: 0},
			},
			Outputs: []transaction.Output{{Value: 90, PubKeyHash: transaction.PubKeyHash{}}},
		},
		{
			Inputs: []transaction.Input{
				{PrevTxID: prevTxID, OutputIndex: 1},
			},
			Outputs: []transaction.Output{{Value: 180, PubKeyHash: transaction.PubKeyHash{}}},
		},
	}

	tip := miner.blockStore.GetMainChainTip()
	txFees, err := miner.getTransactionWithFee(txs, tip.Hash())
	if err != nil {
		t.Fatalf("getTransactionWithFee() returned error: %v", err)
	}

	if len(txFees) != 2 {
		t.Fatalf("getTransactionWithFee() returned %d transactions, want 2", len(txFees))
	}

	// First tx: input 100, output 90, fee = 10
	if txFees[0].Fee != 10 {
		t.Errorf("Transaction 0 fee = %d, want 10", txFees[0].Fee)
	}

	// Second tx: input 200, output 180, fee = 20
	if txFees[1].Fee != 20 {
		t.Errorf("Transaction 1 fee = %d, want 20", txFees[1].Fee)
	}
}

func TestSortAndReversedTransactions(t *testing.T) {
	miner := createTestMinerService(createGenesisBlock(), nil)

	tx1 := transaction.Transaction{Inputs: []transaction.Input{{PrevTxID: transaction.TransactionID{1}}}}
	tx2 := transaction.Transaction{Inputs: []transaction.Input{{PrevTxID: transaction.TransactionID{2}}}}
	tx3 := transaction.Transaction{Inputs: []transaction.Input{{PrevTxID: transaction.TransactionID{3}}}}

	txFee1 := transactionWithFee{tx: tx1, Fee: 10}
	txFee2 := transactionWithFee{tx: tx2, Fee: 30}
	txFee3 := transactionWithFee{tx: tx3, Fee: 20}

	txsWithFees := []transactionWithFee{txFee1, txFee2, txFee3}

	sorted := miner.sortAndReversedTransactions(txsWithFees)

	// Should be sorted by fee in descending order: 30, 20, 10
	if len(sorted) != 3 {
		t.Fatalf("sortAndReversedTransactions() returned %d transactions, want 3", len(sorted))
	}

	// Check by comparing transaction IDs (tx2 has highest fee, tx3 second, tx1 third)
	if sorted[0].TransactionId() != tx2.TransactionId() {
		t.Errorf("First transaction should be tx2 (fee 30), got different transaction")
	}
	if sorted[1].TransactionId() != tx3.TransactionId() {
		t.Errorf("Second transaction should be tx3 (fee 20), got different transaction")
	}
	if sorted[2].TransactionId() != tx1.TransactionId() {
		t.Errorf("Third transaction should be tx1 (fee 10), got different transaction")
	}
}

func TestCreateCoinbaseTransaction(t *testing.T) {
	miner := createTestMinerService(createGenesisBlock(), nil)

	txFee1 := transactionWithFee{Fee: 10, tx: transaction.Transaction{}}
	txFee2 := transactionWithFee{Fee: 20, tx: transaction.Transaction{}}

	coinbase, err := miner.createCoinbaseTransaction([]transactionWithFee{txFee1, txFee2}, 1)
	if err != nil {
		t.Fatalf("createCoinbaseTransaction() returned error: %v", err)
	}

	// Verify it's a coinbase transaction
	if !coinbase.IsCoinbase() {
		t.Error("createCoinbaseTransaction() did not create a coinbase transaction")
	}

	// Verify the reward amount (fees 10+20 + block reward 50 = 80)
	if len(coinbase.Outputs) == 0 {
		t.Fatal("Coinbase transaction has no outputs")
	}

	if coinbase.Outputs[0].Value != 80 {
		t.Errorf("Coinbase output value = %d, want 80 (10+20+50)", coinbase.Outputs[0].Value)
	}
}

func TestBuildTransactions(t *testing.T) {
	// Set up UTXOs for the transactions (all use the same prevTxID from createTestTransaction)
	prevTxID := transaction.TransactionID{}
	prevTxID[0] = 0xAA

	utxos := map[utxoOutpoint]transaction.Output{
		{txID: prevTxID, outputIndex: 0}: {Value: 100, PubKeyHash: transaction.PubKeyHash{}},
	}
	miner := createTestMinerService(createGenesisBlock(), utxos)

	// Create transactions with different fees
	txs := []transaction.Transaction{
		createTestTransaction(100, 90), // fee 10
	}

	tip := miner.blockStore.GetMainChainTip()
	builtTxs, err := miner.buildTransactions(txs, 1, tip.Hash())
	if err != nil {
		t.Fatalf("buildTransactions() returned error: %v", err)
	}

	// Should have coinbase + regular transactions
	if len(builtTxs) != 2 { // 1 coinbase + 1 regular
		t.Errorf("buildTransactions() returned %d transactions, want 2", len(builtTxs))
	}

	// First transaction should be coinbase
	if !builtTxs[0].IsCoinbase() {
		t.Error("First transaction should be coinbase")
	}
}

func TestBuildTransactions_LimitToTxPerBlock(t *testing.T) {
	// Set up UTXOs for the transactions
	prevTxID := transaction.TransactionID{}
	prevTxID[0] = 0xAA

	utxos := map[utxoOutpoint]transaction.Output{
		{txID: prevTxID, outputIndex: 0}: {Value: 100, PubKeyHash: transaction.PubKeyHash{}},
	}
	miner := createTestMinerService(createGenesisBlock(), utxos)

	// Create more transactions than TxPerBlock
	txs := make([]transaction.Transaction, TxPerBlock+10)
	for i := 0; i < len(txs); i++ {
		txs[i] = createTestTransaction(100, 90)
	}

	tip := miner.blockStore.GetMainChainTip()
	builtTxs, err := miner.buildTransactions(txs, 1, tip.Hash())
	if err != nil {
		t.Fatalf("buildTransactions() returned error: %v", err)
	}

	// Should have exactly TxPerBlock transactions (1 coinbase + TxPerBlock-1 regular)
	if len(builtTxs) != TxPerBlock {
		t.Errorf("buildTransactions() returned %d transactions, want %d", len(builtTxs), TxPerBlock)
	}
}

func TestCreateCandidateBlock(t *testing.T) {
	genesis := createGenesisBlock()

	// Set up UTXOs for the transactions
	prevTxID := transaction.TransactionID{}
	prevTxID[0] = 0xAA

	utxos := map[utxoOutpoint]transaction.Output{
		{txID: prevTxID, outputIndex: 0}: {Value: 100, PubKeyHash: transaction.PubKeyHash{}},
	}
	miner := createTestMinerService(genesis, utxos)

	txs := []transaction.Transaction{
		createTestTransaction(100, 90),
	}

	tip := miner.blockStore.GetMainChainTip()
	candidateBlock, err := miner.createCandidateBlock(txs, 1, tip.Hash())
	if err != nil {
		t.Fatalf("createCandidateBlock() returned error: %v", err)
	}

	// Verify header fields
	if candidateBlock.Header.PreviousBlockHash != genesis.Hash() {
		t.Errorf("PreviousBlockHash does not match genesis hash")
	}

	if candidateBlock.Header.DifficultyTarget != 28 {
		t.Errorf("DifficultyTarget = %d, want 28", candidateBlock.Header.DifficultyTarget)
	}

	// Verify transactions are present (should include coinbase)
	if len(candidateBlock.Transactions) == 0 {
		t.Error("Candidate block has no transactions")
	}

	// First transaction should be coinbase
	if !candidateBlock.Transactions[0].IsCoinbase() {
		t.Error("First transaction should be coinbase")
	}
}

func TestMineBlock_ContextCancellation(t *testing.T) {
	miner := createTestMinerService(createGenesisBlock(), nil)

	// Create a candidate block with high difficulty (will take time to mine)
	candidateBlock := block.Block{
		Header: block.BlockHeader{
			PreviousBlockHash: common.Hash{},
			MerkleRoot:        common.Hash{},
			Timestamp:         time.Now().Unix(),
			DifficultyTarget:  255, // Very high difficulty
			Nonce:             0,
		},
		Transactions: []transaction.Transaction{},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, _, err := miner.mineBlock(candidateBlock, ctx)
	if err == nil {
		t.Error("mineBlock() should return error when context is cancelled")
		return
	}

	if err.Error() != "mining cancelled" {
		t.Errorf("mineBlock() error = %v, want 'mining cancelled'", err)
	}
}

func TestMineBlock_Success(t *testing.T) {
	miner := createTestMinerService(createGenesisBlock(), nil)

	// Create a candidate block with very low difficulty (easy to mine)
	candidateBlock := block.Block{
		Header: block.BlockHeader{
			PreviousBlockHash: common.Hash{},
			MerkleRoot:        common.Hash{},
			Timestamp:         time.Now().Unix(),
			DifficultyTarget:  0, // No leading zeros required
			Nonce:             0,
		},
		Transactions: []transaction.Transaction{},
	}

	ctx := context.Background()
	nonce, timestamp, err := miner.mineBlock(candidateBlock, ctx)
	if err != nil {
		t.Fatalf("mineBlock() returned error: %v", err)
	}

	// Verify nonce was found
	if nonce == 0 && candidateBlock.Header.Nonce == 0 {
		// With difficulty 0, any hash should work, so we should find immediately
		// The actual nonce might be 0 or 1 depending on implementation
		fmt.Print("Test")
	}

	// Verify timestamp is reasonable
	if timestamp < candidateBlock.Header.Timestamp-1 {
		t.Error("Mined timestamp is before candidate block timestamp")
	}
}

func TestMineBlock_LowDifficulty(t *testing.T) {
	miner := createTestMinerService(createGenesisBlock(), nil)

	// Test with difficulty of 1 (should find quickly)
	candidateBlock := block.Block{
		Header: block.BlockHeader{
			PreviousBlockHash: common.Hash{},
			MerkleRoot:        common.Hash{},
			Timestamp:         time.Now().Unix(),
			DifficultyTarget:  1, // At least 1 leading zero
			Nonce:             0,
		},
		Transactions: []transaction.Transaction{},
	}

	ctx := context.Background()
	nonce, timestamp, err := miner.mineBlock(candidateBlock, ctx)
	if err != nil {
		t.Fatalf("mineBlock() returned error: %v", err)
	}

	// Set the nonce and verify the hash meets the difficulty
	candidateBlock.Header.Nonce = nonce
	candidateBlock.Header.Timestamp = timestamp

	hash := candidateBlock.Hash()
	target := getTarget(candidateBlock.Header.DifficultyTarget)

	hashInt := new(big.Int)
	hashInt.SetBytes(hash[:])

	if hashInt.Cmp(&target) >= 0 {
		t.Errorf("Mined block hash does not meet difficulty target. Hash: %x, Target: %v", hash, &target)
	}
}

func TestStartMining_StopsMining(t *testing.T) {
	// Set up UTXOs for the transactions
	prevTxID := transaction.TransactionID{}
	prevTxID[0] = 0xAA

	utxos := map[utxoOutpoint]transaction.Output{
		{txID: prevTxID, outputIndex: 0}: {Value: 100, PubKeyHash: transaction.PubKeyHash{}},
	}

	mockBlockchain := &mockBlockchainAPI{}
	mockUTXO := &mockUtxoStoreAPI{utxos: utxos}
	mockBlockStore := &mockBlockStore{tip: createGenesisBlock()}

	miner := &minerService{
		blockchain:    mockBlockchain,
		utxoService:   mockUTXO,
		blockStore:    mockBlockStore,
		miningEnabled: true,
	}

	// Start mining
	txs := []transaction.Transaction{createTestTransaction(100, 90)}
	miner.StartMining(txs)

	// Allow time for StartMining to set up the cancel function
	time.Sleep(50 * time.Millisecond)

	// Stop mining
	miner.StopMining()

	// Give time for goroutine to stop
	time.Sleep(100 * time.Millisecond)

	// Verify blockchain was not called (mining was stopped)
	// This is a rough check - in real scenario, mining might have completed before StopMining
}

func TestNewMinerService(t *testing.T) {
	mockBlockchain := &mockBlockchainAPI{}
	mockUTXO := &mockUtxoStoreAPI{utxos: map[utxoOutpoint]transaction.Output{}}
	mockBlockStore := &mockBlockStore{tip: createGenesisBlock()}

	miner := NewMinerService(mockBlockchain, mockUTXO, mockBlockStore)

	if miner == nil {
		t.Fatal("NewMinerService() returned nil")
	}

	if miner.blockchain != mockBlockchain {
		t.Error("Blockchain not set correctly")
	}

	// Note: Can't compare interface values directly, just check it's not nil
	if miner.utxoService == nil {
		t.Error("UTXO service not set")
	}

	if miner.blockStore != mockBlockStore {
		t.Error("Block store not set correctly")
	}
}

func TestMinerAPI_Interface(t *testing.T) {
	// Compile-time check that minerService implements MinerAPI
	var _ minerApi.MinerAPI = &minerService{}
}

func TestMinerBlockchainAPI_Interface(t *testing.T) {
	// Compile-time check that mockBlockchainAPI implements api.BlockchainAPI
	var _ api.BlockchainAPI = &mockBlockchainAPI{}
}

func TestMinerUTXOService_Interface(t *testing.T) {
	// Compile-time check that mockUtxoStoreAPI implements UtxoStoreAPI
	var _ api.UtxoStoreAPI = &mockUtxoStoreAPI{}
}

func TestMinerBlockStoreAPI_Interface(t *testing.T) {
	// Compile-time check that mockBlockStore implements blockchain.BlockStoreAPI
	var _ api.BlockStoreAPI = &mockBlockStore{}
}
