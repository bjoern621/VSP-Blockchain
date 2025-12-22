package block

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
)

// Block represents a block in the blockchain.
// It consists of a BlockHeader and a list of Transactions.
type Block struct {
	Header       BlockHeader
	Transactions []transaction.Transaction
}
