package block

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"

	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
)

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

// Hash computes the double SHA-256 hash of the block header.
func (h *BlockHeader) Hash() common.Hash {
	var buffer = make([]byte, 0)
	buffer = append(buffer, h.PreviousBlockHash[:]...)
	buffer = append(buffer, h.MerkleRoot[:]...)
	buffer = binary.LittleEndian.AppendUint64(buffer, uint64(h.Timestamp))
	buffer = binary.LittleEndian.AppendUint32(buffer, h.Nonce)
	buffer = append(buffer, h.DifficultyTarget)

	first := sha256.Sum256(buffer)
	second := sha256.Sum256(first[:])
	var hash common.Hash
	copy(hash[:], second[:])
	return hash
}

func (h *BlockHeader) String() string {
	hash := h.Hash()
	return hex.EncodeToString(hash[:])
}
