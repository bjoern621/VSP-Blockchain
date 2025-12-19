package block

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
)

type MerkleProof struct {
	Transaction transaction.Transaction
	Siblings    []common.Hash
	Index       uint32
}
