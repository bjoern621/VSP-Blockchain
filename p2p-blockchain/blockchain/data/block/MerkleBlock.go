package block

import "s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api/blockchain/dto"

type MerkleBlock struct {
	BlockHeader BlockHeader
	Proofs      []MerkleProof
}

func NewMerkleBlockFromDTO(m dto.MerkleBlockDTO) MerkleBlock {
	proofs := make([]MerkleProof, 0, len(m.Proofs))
	for i := range m.Proofs {
		proofs = append(proofs, NewMerkleProofFromDTO(m.Proofs[i]))
	}
	return MerkleBlock{
		BlockHeader: NewBlockHeaderFromDTO(m.Header),
		Proofs:      proofs,
	}
}
