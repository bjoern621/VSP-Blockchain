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
	if !isCoinBaseTransaction(block.Transactions[0]) {
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

func (bvs *BlockValidationService) FullValidation(block block.Block) (bool, error) {
	//TODO
	return true, nil
}

func isCoinBaseTransaction(transaction transaction.Transaction) bool {
	// TODO
	return false
}

func hashSmallerThanTarget(block block.Block) bool {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-block.Header.DifficultyTarget))

	hash := block.Hash()
	var intHash big.Int
	intHash.SetBytes(hash[:])

	return target.Cmp(&intHash) == -1
}
