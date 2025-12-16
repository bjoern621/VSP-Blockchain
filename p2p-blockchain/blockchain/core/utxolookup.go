package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/transaction"
)

type UTXOLookupService interface {
	// GetUTXO returns the output referenced by txID and outputIndex
	GetUTXO(txID transaction.TransactionID, outputIndex uint32) (transaction.Output, bool)
}
