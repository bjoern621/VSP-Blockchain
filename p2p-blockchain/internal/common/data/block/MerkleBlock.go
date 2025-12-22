package block

type MerkleBlock struct {
	BlockHeader BlockHeader
	Proofs      []MerkleProof
}
