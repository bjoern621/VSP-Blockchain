package block

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
)

type MerkleProof struct {
	Transaction transaction.Transaction
	Siblings    []Hash
	Index       uint32
}
