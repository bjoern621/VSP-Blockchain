package validation

import (
	"bytes"
	"fmt"
	"math/big"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"time"
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
	// ValidateHeader Validates the block header (proof of work, timestamp, etc.)
	ValidateHeader(block block.Block) (bool, error)
	// ValidateHeaderOnly Validates a standalone block header (proof of work, timestamp)
	ValidateHeaderOnly(header block.BlockHeader) (bool, error)
	// FullValidation Comprehensive validation including transactions and UTXO set
	FullValidation(block block.Block) (bool, error)
}

type BlockValidationService struct {
}

func NewBlockValidationService() *BlockValidationService {
	return &BlockValidationService{}
}

func (bvs *BlockValidationService) SanityCheck(block block.Block) (bool, error) {
	if len(block.Transactions) < 1 {
		return false, fmt.Errorf("block must contain at least one transaction")
	}
	if block.Transactions[0].IsCoinbase() {
		return false, fmt.Errorf("first transaction must be the coinbase transaction")
	}

	return true, nil
}

func (bvs *BlockValidationService) ValidateHeader(block block.Block) (bool, error) {
	return bvs.ValidateHeaderOnly(block.Header)
}

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

func (bvs *BlockValidationService) FullValidation(block block.Block) (bool, error) {
	//TODO
	/*
		Transaction validation
			- For every transaction:
				- Signature checks
				- Witness validation
				- No double-spends within block
				- No coinbase misuse
		UTXO validation
			- Inputs reference unspent outputs
			- Values are in range
			- Fees are non-negative
			- Coinbase reward is correct
	*/

	if !isMerkleRootValid(block) {
		return false, fmt.Errorf("merkle root in header does not match calculated merkle root")
	}
	return true, nil
}

func isMerkleRootValid(b block.Block) bool {
	merkleRoot := b.MerkleRoot()
	return bytes.Equal(merkleRoot[:], b.Header.MerkleRoot[:])
}
