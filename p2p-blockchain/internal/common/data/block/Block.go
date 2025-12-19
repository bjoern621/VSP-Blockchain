package block

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
)

type Block struct {
	Header       BlockHeader
	Transactions []transaction.Transaction
}
