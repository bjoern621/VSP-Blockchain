package validation

import (
	"errors"
	"testing"

	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
)

// MockUtxoStore is a mock implementation of utxo.UtxoStoreAPI for testing.
type MockUtxoStore struct {
	// utxos maps outpoint keys to outputs for testing
	utxos map[string]transaction.Output
	// validateResult controls the return value of ValidateTransactionFromBlock
	validateResult bool
	// getUtxoError if set, GetUtxoFromBlock will return this error
	getUtxoError error
}

func (m *MockUtxoStore) ValidateTransactionsOfBlock(_ block.Block) bool {
	return true
}

// makeOutpointKey creates a unique key for an outpoint.
func makeOutpointKey(txID transaction.TransactionID, outputIndex uint32) string {
	return string(txID[:]) + string(rune(outputIndex))
}

// InitializeGenesisPool is a mock implementation.
func (m *MockUtxoStore) InitializeGenesisPool(_ block.Block) error {
	return nil
}

// AddNewBlock is a mock implementation.
func (m *MockUtxoStore) AddNewBlock(_ block.Block) error {
	return nil
}

// GetUtxoFromBlock returns a mock UTXO based on the stored utxos map.
func (m *MockUtxoStore) GetUtxoFromBlock(prevTxID transaction.TransactionID, outputIndex uint32, _ common.Hash) (transaction.Output, error) {
	if m.getUtxoError != nil {
		return transaction.Output{}, m.getUtxoError
	}
	key := makeOutpointKey(prevTxID, outputIndex)
	if output, exists := m.utxos[key]; exists {
		return output, nil
	}
	return transaction.Output{}, errors.New("UTXO not found")
}

// ValidateTransactionFromBlock returns the configured validateResult.
func (m *MockUtxoStore) ValidateTransactionFromBlock(_ transaction.Transaction, _ common.Hash) bool {
	return m.validateResult
}

// GetUtxosByPubKeyHashFromBlock is a mock implementation.
func (m *MockUtxoStore) GetUtxosByPubKeyHashFromBlock(_ transaction.PubKeyHash, _ common.Hash) ([]transaction.UTXO, error) {
	return nil, nil
}

// createTestPubKeyAndHash creates a test public key and its corresponding hash.
func createTestPubKeyAndHash() (transaction.PubKey, transaction.PubKeyHash) {
	var pubKey transaction.PubKey
	for i := range pubKey {
		pubKey[i] = byte(i)
	}
	pubKeyHash := transaction.Hash160(pubKey)
	return pubKey, pubKeyHash
}

// createTestTransactionID creates a test transaction ID.
func createTestTransactionID(seed byte) transaction.TransactionID {
	var txID transaction.TransactionID
	for i := range txID {
		txID[i] = seed + byte(i)
	}
	return txID
}

// TestValidateTransaction_Coinbase tests that coinbase transactions are always valid.
func TestValidateTransaction_Coinbase(t *testing.T) {
	mockStore := &MockUtxoStore{
		utxos:          make(map[string]transaction.Output),
		validateResult: false, // Even with false, coinbase should pass
	}
	validator := NewTransactionValidator(mockStore)

	var zeroPubKeyHash transaction.PubKeyHash
	coinbaseTx := transaction.NewCoinbaseTransaction(zeroPubKeyHash, 50, 1)

	valid, err := validator.ValidateTransaction(coinbaseTx, common.Hash{})
	if err != nil {
		t.Errorf("Coinbase validation returned error: %v", err)
	}
	if !valid {
		t.Error("Coinbase transaction should be valid")
	}
}

// TestValidateTransaction_NoInputs tests that transactions without inputs are rejected.
func TestValidateTransaction_NoInputs(t *testing.T) {
	mockStore := &MockUtxoStore{
		utxos:          make(map[string]transaction.Output),
		validateResult: true,
	}
	validator := NewTransactionValidator(mockStore)

	tx := transaction.Transaction{
		Inputs: []transaction.Input{},
		Outputs: []transaction.Output{
			{Value: 100, PubKeyHash: transaction.PubKeyHash{}},
		},
	}

	valid, err := validator.ValidateTransaction(tx, common.Hash{})
	if !errors.Is(err, ErrNoInputs) {
		t.Errorf("Expected ErrNoInputs, got: %v", err)
	}
	if valid {
		t.Error("Transaction with no inputs should be invalid")
	}
}

// TestValidateTransaction_NoOutputs tests that transactions without outputs are rejected.
func TestValidateTransaction_NoOutputs(t *testing.T) {
	mockStore := &MockUtxoStore{
		utxos:          make(map[string]transaction.Output),
		validateResult: true,
	}
	validator := NewTransactionValidator(mockStore)

	pubKey, _ := createTestPubKeyAndHash()
	txID := createTestTransactionID(1)

	tx := transaction.Transaction{
		Inputs: []transaction.Input{
			{
				PrevTxID:    txID,
				OutputIndex: 0,
				PubKey:      pubKey,
			},
		},
		Outputs: []transaction.Output{},
	}

	valid, err := validator.ValidateTransaction(tx, common.Hash{})
	if !errors.Is(err, ErrNoOutputs) {
		t.Errorf("Expected ErrNoOutputs, got: %v", err)
	}
	if valid {
		t.Error("Transaction with no outputs should be invalid")
	}
}

// TestValidateTransaction_UTXONotFound tests that transactions with non-existent UTXOs are rejected.
func TestValidateTransaction_UTXONotFound(t *testing.T) {
	mockStore := &MockUtxoStore{
		utxos:          make(map[string]transaction.Output),
		validateResult: false, // UTXO validation fails
	}
	validator := NewTransactionValidator(mockStore)

	pubKey, _ := createTestPubKeyAndHash()
	txID := createTestTransactionID(1)

	tx := transaction.Transaction{
		Inputs: []transaction.Input{
			{
				PrevTxID:    txID,
				OutputIndex: 0,
				PubKey:      pubKey,
			},
		},
		Outputs: []transaction.Output{
			{Value: 50, PubKeyHash: transaction.PubKeyHash{}},
		},
	}

	valid, err := validator.ValidateTransaction(tx, common.Hash{})
	if !errors.Is(err, ErrUTXONotFound) {
		t.Errorf("Expected ErrUTXONotFound, got: %v", err)
	}
	if valid {
		t.Error("Transaction referencing non-existent UTXO should be invalid")
	}
}

// TestValidateTransaction_PubKeyHashMismatch tests that transactions with wrong public key are rejected.
func TestValidateTransaction_PubKeyHashMismatch(t *testing.T) {
	pubKey, _ := createTestPubKeyAndHash()
	txID := createTestTransactionID(1)

	// Create a different PubKeyHash that doesn't match the input's pubkey
	var differentPubKeyHash transaction.PubKeyHash
	for i := range differentPubKeyHash {
		differentPubKeyHash[i] = byte(i + 100) // Different from Hash160(pubKey)
	}

	mockStore := &MockUtxoStore{
		utxos: map[string]transaction.Output{
			makeOutpointKey(txID, 0): {
				Value:      100,
				PubKeyHash: differentPubKeyHash, // Doesn't match input's pubkey
			},
		},
		validateResult: true,
	}
	validator := NewTransactionValidator(mockStore)

	tx := transaction.Transaction{
		Inputs: []transaction.Input{
			{
				PrevTxID:    txID,
				OutputIndex: 0,
				PubKey:      pubKey,
			},
		},
		Outputs: []transaction.Output{
			{Value: 50, PubKeyHash: transaction.PubKeyHash{}},
		},
	}

	valid, err := validator.ValidateTransaction(tx, common.Hash{})
	if !errors.Is(err, ErrPubKeyHashMismatch) {
		t.Errorf("Expected ErrPubKeyHashMismatch, got: %v", err)
	}
	if valid {
		t.Error("Transaction with mismatched public key hash should be invalid")
	}
}

// TestValidateTransaction_DuplicateInput tests that transactions with duplicate inputs are rejected.
func TestValidateTransaction_DuplicateInput(t *testing.T) {
	pubKey, pubKeyHash := createTestPubKeyAndHash()
	txID := createTestTransactionID(1)

	mockStore := &MockUtxoStore{
		utxos: map[string]transaction.Output{
			makeOutpointKey(txID, 0): {
				Value:      100,
				PubKeyHash: pubKeyHash,
			},
		},
		validateResult: true,
	}
	validator := NewTransactionValidator(mockStore)

	// Create a transaction with the same input twice (double-spend attempt)
	tx := transaction.Transaction{
		Inputs: []transaction.Input{
			{
				PrevTxID:    txID,
				OutputIndex: 0,
				PubKey:      pubKey,
			},
			{
				PrevTxID:    txID,
				OutputIndex: 0, // Same outpoint as first input
				PubKey:      pubKey,
			},
		},
		Outputs: []transaction.Output{
			{Value: 50, PubKeyHash: transaction.PubKeyHash{}},
		},
	}

	valid, err := validator.ValidateTransaction(tx, common.Hash{})
	if !errors.Is(err, ErrDuplicateInput) {
		t.Errorf("Expected ErrDuplicateInput, got: %v", err)
	}
	if valid {
		t.Error("Transaction with duplicate inputs should be invalid")
	}
}

// TestValidateTransaction_InsufficientInputs tests that transactions where inputs < outputs are rejected.
func TestValidateTransaction_InsufficientInputs(t *testing.T) {
	pubKey, pubKeyHash := createTestPubKeyAndHash()
	txID := createTestTransactionID(1)

	mockStore := &MockUtxoStore{
		utxos: map[string]transaction.Output{
			makeOutpointKey(txID, 0): {
				Value:      50, // Only 50 available
				PubKeyHash: pubKeyHash,
			},
		},
		validateResult: true,
	}
	validator := NewTransactionValidator(mockStore)

	tx := transaction.Transaction{
		Inputs: []transaction.Input{
			{
				PrevTxID:    txID,
				OutputIndex: 0,
				PubKey:      pubKey,
			},
		},
		Outputs: []transaction.Output{
			{Value: 100, PubKeyHash: transaction.PubKeyHash{}}, // Trying to spend 100
		},
	}

	valid, err := validator.ValidateTransaction(tx, common.Hash{})
	if !errors.Is(err, ErrInsufficientInputs) {
		t.Errorf("Expected ErrInsufficientInputs, got: %v", err)
	}
	if valid {
		t.Error("Transaction with insufficient inputs should be invalid")
	}
}

// TestValidateTransaction_Success tests a valid transaction.
func TestValidateTransaction_Success(t *testing.T) {
	pubKey, pubKeyHash := createTestPubKeyAndHash()
	txID := createTestTransactionID(1)

	mockStore := &MockUtxoStore{
		utxos: map[string]transaction.Output{
			makeOutpointKey(txID, 0): {
				Value:      100,
				PubKeyHash: pubKeyHash,
			},
		},
		validateResult: true,
	}
	validator := NewTransactionValidator(mockStore)

	tx := transaction.Transaction{
		Inputs: []transaction.Input{
			{
				PrevTxID:    txID,
				OutputIndex: 0,
				PubKey:      pubKey,
			},
		},
		Outputs: []transaction.Output{
			{Value: 90, PubKeyHash: transaction.PubKeyHash{}}, // 10 as fee
		},
	}

	valid, err := validator.ValidateTransaction(tx, common.Hash{})
	if err != nil {
		t.Errorf("Valid transaction returned error: %v", err)
	}
	if !valid {
		t.Error("Valid transaction should pass validation")
	}
}

// TestValidateTransaction_ExactInputsMatchOutputs tests a transaction where inputs exactly match outputs.
func TestValidateTransaction_ExactInputsMatchOutputs(t *testing.T) {
	pubKey, pubKeyHash := createTestPubKeyAndHash()
	txID := createTestTransactionID(1)

	mockStore := &MockUtxoStore{
		utxos: map[string]transaction.Output{
			makeOutpointKey(txID, 0): {
				Value:      100,
				PubKeyHash: pubKeyHash,
			},
		},
		validateResult: true,
	}
	validator := NewTransactionValidator(mockStore)

	tx := transaction.Transaction{
		Inputs: []transaction.Input{
			{
				PrevTxID:    txID,
				OutputIndex: 0,
				PubKey:      pubKey,
			},
		},
		Outputs: []transaction.Output{
			{Value: 100, PubKeyHash: transaction.PubKeyHash{}}, // Exact match, 0 fee
		},
	}

	valid, err := validator.ValidateTransaction(tx, common.Hash{})
	if err != nil {
		t.Errorf("Transaction with exact input/output match returned error: %v", err)
	}
	if !valid {
		t.Error("Transaction with exact input/output match should be valid")
	}
}

// TestValidateTransaction_MultipleInputsAndOutputs tests a transaction with multiple inputs and outputs.
func TestValidateTransaction_MultipleInputsAndOutputs(t *testing.T) {
	pubKey, pubKeyHash := createTestPubKeyAndHash()
	txID1 := createTestTransactionID(1)
	txID2 := createTestTransactionID(2)

	mockStore := &MockUtxoStore{
		utxos: map[string]transaction.Output{
			makeOutpointKey(txID1, 0): {
				Value:      50,
				PubKeyHash: pubKeyHash,
			},
			makeOutpointKey(txID2, 0): {
				Value:      60,
				PubKeyHash: pubKeyHash,
			},
		},
		validateResult: true,
	}
	validator := NewTransactionValidator(mockStore)

	tx := transaction.Transaction{
		Inputs: []transaction.Input{
			{
				PrevTxID:    txID1,
				OutputIndex: 0,
				PubKey:      pubKey,
			},
			{
				PrevTxID:    txID2,
				OutputIndex: 0,
				PubKey:      pubKey,
			},
		},
		Outputs: []transaction.Output{
			{Value: 40, PubKeyHash: transaction.PubKeyHash{}},
			{Value: 50, PubKeyHash: transaction.PubKeyHash{}},
			// Total outputs: 90, inputs: 110, fee: 20
		},
	}

	valid, err := validator.ValidateTransaction(tx, common.Hash{})
	if err != nil {
		t.Errorf("Valid multi-input/output transaction returned error: %v", err)
	}
	if !valid {
		t.Error("Valid multi-input/output transaction should pass validation")
	}
}

// TestTransactionValidator_CreateOutpointKey tests that the outpoint key creation is consistent.
func TestTransactionValidator_CreateOutpointKey(t *testing.T) {
	mockStore := &MockUtxoStore{}
	validator := &TransactionValidator{utxoStore: mockStore}

	txID := createTestTransactionID(1)
	input := transaction.Input{
		PrevTxID:    txID,
		OutputIndex: 5,
	}

	key1 := validator.createOutpointKey(input)
	key2 := validator.createOutpointKey(input)

	if key1 != key2 {
		t.Error("createOutpointKey should return consistent keys for the same input")
	}

	// Different output index should produce different key
	input2 := transaction.Input{
		PrevTxID:    txID,
		OutputIndex: 6,
	}
	key3 := validator.createOutpointKey(input2)

	if key1 == key3 {
		t.Error("createOutpointKey should return different keys for different output indices")
	}
}

// TestTransactionValidator_CalculateOutputSum tests the output sum calculation.
func TestTransactionValidator_CalculateOutputSum(t *testing.T) {
	mockStore := &MockUtxoStore{}
	validator := &TransactionValidator{utxoStore: mockStore}

	tests := []struct {
		name     string
		outputs  []transaction.Output
		expected uint64
	}{
		{
			name:     "empty outputs",
			outputs:  []transaction.Output{},
			expected: 0,
		},
		{
			name: "single output",
			outputs: []transaction.Output{
				{Value: 100},
			},
			expected: 100,
		},
		{
			name: "multiple outputs",
			outputs: []transaction.Output{
				{Value: 100},
				{Value: 200},
				{Value: 50},
			},
			expected: 350,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := transaction.Transaction{Outputs: tt.outputs}
			result := validator.calculateOutputSum(tx)
			if result != tt.expected {
				t.Errorf("calculateOutputSum() = %d, want %d", result, tt.expected)
			}
		})
	}
}

// TestTransactionValidator_CheckDuplicateInput tests duplicate input detection.
func TestTransactionValidator_CheckDuplicateInput(t *testing.T) {
	mockStore := &MockUtxoStore{}
	validator := &TransactionValidator{utxoStore: mockStore}

	seenOutpoints := make(map[string]struct{})

	// First occurrence should not be a duplicate
	err := validator.checkDuplicateInput("key1", seenOutpoints)
	if err != nil {
		t.Errorf("First occurrence should not be duplicate, got: %v", err)
	}
	seenOutpoints["key1"] = struct{}{}

	// Second occurrence should be a duplicate
	err = validator.checkDuplicateInput("key1", seenOutpoints)
	if !errors.Is(err, ErrDuplicateInput) {
		t.Errorf("Expected ErrDuplicateInput, got: %v", err)
	}

	// Different key should not be a duplicate
	err = validator.checkDuplicateInput("key2", seenOutpoints)
	if err != nil {
		t.Errorf("Different key should not be duplicate, got: %v", err)
	}
}

// TestTransactionValidator_ValidateBasicStructure tests basic structure validation.
func TestTransactionValidator_ValidateBasicStructure(t *testing.T) {
	mockStore := &MockUtxoStore{}
	validator := &TransactionValidator{utxoStore: mockStore}

	tests := []struct {
		name        string
		tx          transaction.Transaction
		expectedErr error
	}{
		{
			name: "valid structure",
			tx: transaction.Transaction{
				Inputs:  []transaction.Input{{PrevTxID: transaction.TransactionID{1}}},
				Outputs: []transaction.Output{{Value: 100}},
			},
			expectedErr: nil,
		},
		{
			name: "no inputs",
			tx: transaction.Transaction{
				Inputs:  []transaction.Input{},
				Outputs: []transaction.Output{{Value: 100}},
			},
			expectedErr: ErrNoInputs,
		},
		{
			name: "no outputs",
			tx: transaction.Transaction{
				Inputs:  []transaction.Input{{PrevTxID: transaction.TransactionID{1}}},
				Outputs: []transaction.Output{},
			},
			expectedErr: ErrNoOutputs,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateBasicStructure(tt.tx)
			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("validateBasicStructure() error = %v, want %v", err, tt.expectedErr)
			}
		})
	}
}

// TestTransactionValidator_ValidateValueConservation tests value conservation validation.
func TestTransactionValidator_ValidateValueConservation(t *testing.T) {
	mockStore := &MockUtxoStore{}
	validator := &TransactionValidator{utxoStore: mockStore}

	tests := []struct {
		name        string
		tx          transaction.Transaction
		inputSum    uint64
		expectedErr error
	}{
		{
			name: "inputs greater than outputs (with fee)",
			tx: transaction.Transaction{
				Outputs: []transaction.Output{{Value: 90}},
			},
			inputSum:    100,
			expectedErr: nil,
		},
		{
			name: "inputs equal to outputs (no fee)",
			tx: transaction.Transaction{
				Outputs: []transaction.Output{{Value: 100}},
			},
			inputSum:    100,
			expectedErr: nil,
		},
		{
			name: "inputs less than outputs",
			tx: transaction.Transaction{
				Outputs: []transaction.Output{{Value: 150}},
			},
			inputSum:    100,
			expectedErr: ErrInsufficientInputs,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateValueConservation(tt.tx, tt.inputSum)
			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("validateValueConservation() error = %v, want %v", err, tt.expectedErr)
			}
		})
	}
}

// TestTransactionValidator_ValidatePubKeyHash tests public key hash validation.
func TestTransactionValidator_ValidatePubKeyHash(t *testing.T) {
	mockStore := &MockUtxoStore{}
	validator := &TransactionValidator{utxoStore: mockStore}

	pubKey, pubKeyHash := createTestPubKeyAndHash()

	tests := []struct {
		name        string
		input       transaction.Input
		output      transaction.Output
		expectedErr error
	}{
		{
			name:        "matching public key hash",
			input:       transaction.Input{PubKey: pubKey},
			output:      transaction.Output{Value: 100, PubKeyHash: pubKeyHash},
			expectedErr: nil,
		},
		{
			name:  "mismatched public key hash",
			input: transaction.Input{PubKey: pubKey},
			output: transaction.Output{
				Value:      100,
				PubKeyHash: transaction.PubKeyHash{1, 2, 3}, // Different hash
			},
			expectedErr: ErrPubKeyHashMismatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validatePubKeyHash(tt.input, tt.output)
			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("validatePubKeyHash() error = %v, want %v", err, tt.expectedErr)
			}
		})
	}
}
