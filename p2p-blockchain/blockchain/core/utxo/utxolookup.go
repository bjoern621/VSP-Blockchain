package utxo

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
)

type UTXOLookupService interface {
	// GetUTXO returns the output referenced by txID and outputIndex
	GetUTXO(txID transaction.TransactionID, outputIndex uint32) (transaction.Output, bool)
}

type UTXOLookupImpl struct{}

func (u *UTXOLookupImpl) GetUTXO(txID transaction.TransactionID, outputIndex uint32) (transaction.Output, bool) {
	// TODO
	return transaction.Output{}, false
}
