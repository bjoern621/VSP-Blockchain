package block

import "s3b/vsp-blockchain/p2p-blockchain/internal/common"

type BlockHeader struct {
	Hash             common.Hash
	MerkleRoot       common.Hash
	Timestamp        int64
	DifficultyTarget uint32
	Nonce            uint32
}
