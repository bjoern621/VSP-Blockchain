package validation

import (
	"bytes"
	"fmt"
	"math/big"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/utxo"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
)

const minutesAheadLimit = 5

// BlockValidationAPI defines the interface for block validation
// There are three levels of validation:
//  1. Sanity Check: Basic checks on the block structure and content
//  2. Header Validation: Validates the block header (proof of work, timestamp, etc.)
//     -> Block can still be invalid but be added to the block store as an orphan
//  3. Full Validation: Comprehensive validation including transactions and UTXO set
//     -> Block must be valid to be added to the main chain
type BlockValidationAPI interface {
	// SanityCheck Sanity Check: Basic checks on the block structure and content
	SanityCheck(block block.Block) (bool, error)
	// ValidateHeaderOnly Validates a standalone block header (proof of work, timestamp)
	ValidateHeaderOnly(header block.BlockHeader) (bool, error)
	// FullValidation Comprehensive validation including transactions and UTXO set
	FullValidation(block block.Block) (bool, error)
}

type BlockValidationService struct {
	txValidator TransactionValidatorAPI
	utxoStore   utxo.UtxoStoreAPI
}

// NewBlockValidationService creates new BlockValidationService
func NewBlockValidationService() *BlockValidationService {
	return &BlockValidationService{}
}

// SetDependencies sets the dependencies for BlockValidationService
func (bvs *BlockValidationService) SetDependencies(
	txValidator TransactionValidatorAPI,
	utxoStore utxo.UtxoStoreAPI,
) {
	bvs.txValidator = txValidator
	bvs.utxoStore = utxoStore
}

// SanityCheck Sanity Check: Basic checks on the block structure and content
func (bvs *BlockValidationService) SanityCheck(block block.Block) (bool, error) {
	if len(block.Transactions) < 1 {
		return false, fmt.Errorf("block must contain at least one transaction")
	}
	if !block.Transactions[0].IsCoinbase() {
		return false, fmt.Errorf("first transaction must be the coinbase transaction")
	}

	return true, nil
}

// ValidateHeaderOnly Validates a standalone block header (proof of work, timestamp)
func (bvs *BlockValidationService) ValidateHeaderOnly(header block.BlockHeader) (bool, error) {
	if !headerHashSmallerThanTarget(header) {
		return false, fmt.Errorf("header hash does not meet difficulty target")
	}
	if headerTimeIsTooFarInFuture(header) {
		return false, fmt.Errorf("header timestamp is too far in the future")
	}

	return true, nil
}

func headerTimeIsTooFarInFuture(h block.BlockHeader) bool {
	currentTime := time.Now()
	limit := currentTime.Add(time.Minute * minutesAheadLimit)

	blockTime := time.Unix(h.Timestamp, 0)
	return blockTime.After(limit)
}

func headerHashSmallerThanTarget(header block.BlockHeader) bool {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-uint(header.DifficultyTarget)))

	hash := header.Hash()
	var intHash big.Int
	intHash.SetBytes(hash[:])

	return intHash.Cmp(target) < 0
}

// FullValidation Comprehensive validation including transactions and UTXO set
func (bvs *BlockValidationService) FullValidation(block block.Block) (bool, error) {
	// Validate merkle root
	if !isMerkleRootValid(block) {
		return false, fmt.Errorf("merkle root in header does not match calculated merkle root")
	}

	usedInputs := mapset.NewSet[string]()
	var coinbaseFound = false
	prevBlockHash := block.Header.PreviousBlockHash

	for i, tx := range block.Transactions {
		if tx.IsCoinbase() {
			if coinbaseFound {
				return false, fmt.Errorf("multiple coinbase transactions in block %v", block.Header.Hash())
			}

			coinbaseFound = true
			continue
		}

		// doublespending within block
		for _, input := range tx.Inputs {
			key := createInputKey(input)
			if usedInputs.Contains(key) {
				return false, fmt.Errorf("double-spend detected within block %v for tx %v", block.Header.Hash(), tx.Hash())
			}
			usedInputs.Add(key)
		}

		valid, err := bvs.txValidator.ValidateTransaction(tx, prevBlockHash)
		if err != nil {
			return false, fmt.Errorf("transaction %d validation failed: %w", i, err)
		}
		if !valid {
			return false, fmt.Errorf("transaction %d is invalid", i)
		}

		// Verify all signatures in the transaction
		valid, err = bvs.verifyTransactionSignatures(tx, prevBlockHash)
		if err != nil {
			return false, fmt.Errorf("signature verification failed for transaction %d: %w", i, err)
		}
		if !valid {
			return false, fmt.Errorf("invalid transaction signature in block %v for tx %v", block.Header.Hash(), tx.Hash())
		}
	}

	return true, nil
}

// verifyTransactionSignatures verifies all signatures in a transaction.
func (bvs *BlockValidationService) verifyTransactionSignatures(tx transaction.Transaction, blockHash common.Hash) (bool, error) {
	// Collect referenced outputs for all inputs
	referencedOutputs := make([]transaction.Output, len(tx.Inputs))

	for i, input := range tx.Inputs {
		output, err := bvs.utxoStore.GetUtxoFromBlock(input.PrevTxID, input.OutputIndex, blockHash)
		if err != nil {
			return false, fmt.Errorf("failed to get UTXO for input %d: %w", i, err)
		}
		referencedOutputs[i] = output
	}

	return tx.VerifyAllSignatures(referencedOutputs)
}

// createInputKey creates a unique key for an input to detect double-spends.
func createInputKey(input transaction.Input) string {
	return string(input.PrevTxID[:]) + string(rune(input.OutputIndex))
}

func isMerkleRootValid(b block.Block) bool {
	merkleRoot := b.MerkleRoot()
	return bytes.Equal(merkleRoot[:], b.Header.MerkleRoot[:])
}
