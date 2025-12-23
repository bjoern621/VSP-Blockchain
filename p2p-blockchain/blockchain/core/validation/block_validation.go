package validation

import "s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"

type BlockValidationAPI interface {
	ValidateBlock(block block.Block) (bool, error)
}

type BlockValidationService struct {
}

func NewBlockValidationService() *BlockValidationService {
	return &BlockValidationService{}
}

func (bvs *BlockValidationService) ValidateBlock(block block.Block) (bool, error) {
	return true, nil
}
