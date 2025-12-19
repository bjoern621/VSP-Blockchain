package block

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/transaction"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api/blockchain/dto"
)

type MerkleProof struct {
	Transaction transaction.Transaction
	Siblings    []Hash
	Index       uint32
}

func NewMerkleProofFromDTO(p dto.MerkleProofDTO) MerkleProof {
	siblings := make([]Hash, 0, len(p.Siblings))
	for i := range p.Siblings {
		siblings = append(siblings, NewHashFromDTO(p.Siblings[i]))
	}
	return MerkleProof{
		Transaction: transaction.NewTransactionFromDTO(p.Transaction),
		Siblings:    siblings,
		Index:       p.Index,
	}
}
