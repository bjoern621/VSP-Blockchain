package utxo

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/transaction"
)

type UTXOLookupService interface {
	// GetUTXO returns the output referenced by txID and outputIndex
	GetUTXO(txID transaction.TransactionID, outputIndex uint32) (transaction.Output, bool)
}

type UTXOLookupImpl struct{}

// TODO
func (u *UTXOLookupImpl) GetUTXO(txID transaction.TransactionID, outputIndex uint32) (transaction.Output, bool) {
	panic("implement me")
	return transaction.Output{}, false
}
