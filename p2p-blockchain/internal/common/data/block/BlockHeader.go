package block

import "s3b/vsp-blockchain/p2p-blockchain/internal/common"

// BlockHeader represents the header of a block in the blockchain.
type BlockHeader struct {
	// The hash of the previous block header in the chain.
	PreviousBlockHash common.Hash
	MerkleRoot        common.Hash
	// The timestamp of the block. In Unix epoch format.
	Timestamp int64
	// DifficultyTarget is the minimum number of leading zero bits required in the block hash [0; 255].
	DifficultyTarget uint8
	// Nonce is a counter used for the proof-of-work algorithm to find a valid hash.
	Nonce uint32
}

const StandardDifficultyTarget uint8 = 28 // ~100 seconds per block
