package block

type BlockHeader struct {
	Hash             Hash
	MerkleRoot       Hash
	Timestamp        int64
	DifficultyTarget uint32
	Nonce            uint32
}
