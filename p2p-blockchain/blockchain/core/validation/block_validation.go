package validation

import "s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"

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
	return true, nil
}

func (bvs *BlockValidationService) ValidateHeader(block block.Block) (bool, error) {
	return true, nil
}

func (bvs *BlockValidationService) FullValidation(block block.Block) (bool, error) {
	return true, nil
}
