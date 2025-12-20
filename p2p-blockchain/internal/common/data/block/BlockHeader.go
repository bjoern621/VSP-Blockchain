package block

import "s3b/vsp-blockchain/p2p-blockchain/internal/common"

// BlockHeader represents the header of a block in the blockchain.
type BlockHeader struct {
	PreviousBlockHash common.Hash
	MerkleRoot        common.Hash
	Timestamp         int64
	DifficultyTarget  uint32
	Nonce             uint32
}
