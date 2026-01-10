package validation

import (
	"fmt"
	"math/big"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"time"
)

const minutesAheadLimit = 5

type BlockValidationAPI interface {
	SanityCheck(block block.Block) (bool, error)
	ValidateHeader(block block.Block) (bool, error)
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
	if !hashSmallerThanTarget(block) {
		return false, fmt.Errorf("block hash is not smaller than target")
	}
	if timeIsTooFarInFuture(block) {
		return false, fmt.Errorf("block timestamp is too far in the future")
	}

	return true, nil
}

func timeIsTooFarInFuture(b block.Block) bool {
	currentTime := time.Now()
	limit := currentTime.Add(time.Minute * minutesAheadLimit)

	blockTime := time.Unix(b.Header.Timestamp, 0)
	return blockTime.After(limit)
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
	return merkleRoot.Equals(b.Header.MerkleRoot)
}

func hashSmallerThanTarget(block block.Block) bool {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-uint(block.Header.DifficultyTarget)))

	hash := block.Hash()
	var intHash big.Int
	intHash.SetBytes(hash[:])

	return target.Cmp(&intHash) == -1
}
