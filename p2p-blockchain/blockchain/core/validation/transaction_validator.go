package validation

import (
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

type TransactionValidatorAPI interface {
	ValidateTransaction(tx transaction.Transaction, blockHash common.Hash) (bool, error)
}

type TransactionValidator struct {
	utxoStore utxo.UtxoStoreAPI
}

func NewTransactionValidator(utxoStoreAPI utxo.UtxoStoreAPI) TransactionValidatorAPI {
	return &TransactionValidator{
		utxoStore: utxoStoreAPI,
	}
}

func (t *TransactionValidator) ValidateTransaction(tx transaction.Transaction, blockHash common.Hash) (bool, error) {
	// Step 1: Handle coinbase transactions specially
	// Coinbase transactions create new coins and don't reference UTXOs
	if tx.IsCoinbase() {
		// Coinbase transactions are valid by structure (no UTXO validation needed)
		// Note: Coinbase reward validation is done at block level
		return true, nil
	}

	// Step 2: Basic sanity checks
	if len(tx.Inputs) == 0 {
		return false, ErrNoInputs
	}
	if len(tx.Outputs) == 0 {
		return false, ErrNoOutputs
	}

	// Step 3: Track seen outpoints to detect double-spending within the transaction
	seenOutpoints := make(map[string]struct{})

	var inputSum uint64

	for i, input := range tx.Inputs {
		// Create unique key for this outpoint
		outpointKey := string(input.PrevTxID[:]) + string(rune(input.OutputIndex))

		// Step 4: Check for duplicate inputs (double-spend within same transaction)
		if _, exists := seenOutpoints[outpointKey]; exists {
			return false, ErrDuplicateInput
		}
		seenOutpoints[outpointKey] = struct{}{}

		valid := t.utxoStore.ValidateTransactionFromBlock(tx, blockHash)
		if !valid {
			return false, ErrUTXONotFound
		}

		referencedOutput, err := t.utxoStore.GetUtxoFromBlock(input.PrevTxID, input.OutputIndex, blockHash)
		if err != nil {
			return false, ErrUTXONotFound
		}

		// Step 3: Validate public key hash binding
		// Compute HASH160(Input.PubKey) and compare with the referenced UTXO's PubKeyHash
		computedPubKeyHash := transaction.Hash160(input.PubKey)
		if computedPubKeyHash != referencedOutput.PubKeyHash {
			return false, ErrPubKeyHashMismatch
		}

		// Accumulate input value for step 4
		inputSum += referencedOutput.Value

		_ = i // Input index used for potential future signature validation (step 2)
	}

	// Step 4: Calculate output sum and verify value conservation
	var outputSum uint64
	for _, output := range tx.Outputs {
		outputSum += output.Value
	}

	// Inputs must be >= outputs (difference is the transaction fee)
	if inputSum < outputSum {
		return false, ErrInsufficientInputs
	}

	return true, nil
}
