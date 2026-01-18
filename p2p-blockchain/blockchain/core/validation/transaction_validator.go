package validation

import (
	"encoding/binary"
	"errors"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/utxo"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
)

var (
	ErrUTXONotFound       = errors.New("UTXO not found")
	ErrPubKeyHashMismatch = errors.New("public key hash does not match referenced UTXO")
	ErrDuplicateInput     = errors.New("duplicate input detected (double-spend attempt)")
	ErrInsufficientInputs = errors.New("sum of inputs is less than sum of outputs")
	ErrNoInputs           = errors.New("transaction has no inputs")
	ErrNoOutputs          = errors.New("transaction has no outputs")
)

// TransactionValidatorAPI defines the interface for validating transactions against the UTXO set.
type TransactionValidatorAPI interface {
	// ValidateTransaction validates a transaction against the UTXO set at a specific block.
	// It returns true if the transaction is valid, false otherwise.
	// For coinbase transactions, validation is always successful (reward validation is done at block level).
	ValidateTransaction(tx transaction.Transaction, blockHash common.Hash) (bool, error)
}

// TransactionValidator implements TransactionValidatorAPI and validates transactions
// against the UTXO set stored in the blockchain.
type TransactionValidator struct {
	utxoStore utxo.UtxoStoreAPI
}

// NewTransactionValidator creates a new TransactionValidator with the given UTXO store.
func NewTransactionValidator(utxoStoreAPI utxo.UtxoStoreAPI) TransactionValidatorAPI {
	return &TransactionValidator{
		utxoStore: utxoStoreAPI,
	}
}

// ValidateTransaction validates a transaction against the UTXO set at the specified block.
// It performs the following checks:
//  1. Coinbase transactions are automatically valid (no UTXO validation needed)
//  2. Basic sanity checks (non-empty inputs and outputs)
//  3. Duplicate input detection (prevents double-spending within the same transaction)
//  4. UTXO existence verification
//  5. Public key hash validation (ensures the spender owns the referenced output)
//  6. Value conservation (inputs >= outputs, difference is the transaction fee)
func (t *TransactionValidator) ValidateTransaction(tx transaction.Transaction, blockHash common.Hash) (bool, error) {
	// Coinbase transactions are valid by structure (no UTXO validation needed)
	if tx.IsCoinbase() {
		return true, nil
	}

	if err := t.validateBasicStructure(tx); err != nil {
		return false, err
	}

	inputSum, err := t.validateInputs(tx, blockHash)
	if err != nil {
		return false, err
	}

	if err := t.validateValueConservation(tx, inputSum); err != nil {
		return false, err
	}

	return true, nil
}

// validateBasicStructure performs basic sanity checks on the transaction structure.
// It ensures the transaction has at least one input and one output.
func (t *TransactionValidator) validateBasicStructure(tx transaction.Transaction) error {
	if len(tx.Inputs) == 0 {
		return ErrNoInputs
	}
	if len(tx.Outputs) == 0 {
		return ErrNoOutputs
	}
	return nil
}

// validateInputs validates all inputs in the transaction.
// It checks for duplicate inputs, verifies UTXO existence, and validates public key hash bindings.
// Returns the total sum of input values if successful.
func (t *TransactionValidator) validateInputs(tx transaction.Transaction, blockHash common.Hash) (uint64, error) {
	seenOutpoints := make(map[string]struct{})
	var inputSum uint64

	for _, input := range tx.Inputs {
		outpointKey := t.createOutpointKey(input)

		if err := t.checkDuplicateInput(outpointKey, seenOutpoints); err != nil {
			return 0, err
		}
		seenOutpoints[outpointKey] = struct{}{}

		referencedOutput, err := t.validateAndGetReferencedOutput(tx, input, blockHash)
		if err != nil {
			return 0, err
		}

		if err := t.validatePubKeyHash(input, referencedOutput); err != nil {
			return 0, err
		}

		inputSum += referencedOutput.Value
	}

	return inputSum, nil
}

// createOutpointKey creates a unique string key for an outpoint (transaction ID + output index).
// This is used to detect duplicate inputs within the same transaction.
func (t *TransactionValidator) createOutpointKey(input transaction.Input) string {
	indexBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(indexBytes, input.OutputIndex)
	return string(input.PrevTxID[:]) + string(indexBytes)
}

// checkDuplicateInput checks if the given outpoint has already been seen in this transaction.
// Returns ErrDuplicateInput if the outpoint was already used.
func (t *TransactionValidator) checkDuplicateInput(outpointKey string, seenOutpoints map[string]struct{}) error {
	if _, exists := seenOutpoints[outpointKey]; exists {
		return ErrDuplicateInput
	}
	return nil
}

// validateAndGetReferencedOutput validates that the referenced UTXO exists and returns it.
// Returns ErrUTXONotFound if the UTXO does not exist or the transaction is invalid.
func (t *TransactionValidator) validateAndGetReferencedOutput(tx transaction.Transaction, input transaction.Input, blockHash common.Hash) (transaction.Output, error) {
	referencedOutput, err := t.utxoStore.GetUtxoFromBlock(input.PrevTxID, input.OutputIndex, blockHash)
	if err != nil {
		return transaction.Output{}, ErrUTXONotFound
	}

	return referencedOutput, nil
}

// validatePubKeyHash verifies that the input's public key hashes to the referenced output's public key hash.
// This ensures that the spender actually owns the UTXO they are trying to spend.
func (t *TransactionValidator) validatePubKeyHash(input transaction.Input, referencedOutput transaction.Output) error {
	computedPubKeyHash := transaction.Hash160(input.PubKey)
	if computedPubKeyHash != referencedOutput.PubKeyHash {
		return ErrPubKeyHashMismatch
	}
	return nil
}

// validateValueConservation ensures that the total input value is at least equal to the total output value.
// The difference between inputs and outputs is the transaction fee.
func (t *TransactionValidator) validateValueConservation(tx transaction.Transaction, inputSum uint64) error {
	outputSum := t.calculateOutputSum(tx)

	if inputSum < outputSum {
		return ErrInsufficientInputs
	}
	return nil
}

// calculateOutputSum computes the total value of all outputs in the transaction.
func (t *TransactionValidator) calculateOutputSum(tx transaction.Transaction) uint64 {
	var outputSum uint64
	for _, output := range tx.Outputs {
		outputSum += output.Value
	}
	return outputSum
}
