package validation

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/utxo"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
)

type TransactionValidatorAPI interface {
	ValidateTransaction(tx transaction.Transaction, blockHash common.Hash) (bool, error)
}

type TransactionValidator struct {
	// dependencies like UTXO store, signature verifier, etc.
}

func NewTransactionValidator(api utxo.UtxoStoreAPI) TransactionValidatorAPI {
	return &TransactionValidator{
		// initialize dependencies
	}
}

func (t *TransactionValidator) ValidateTransaction(tx transaction.Transaction, blockHash common.Hash) (bool, error) {
	//TODO implement me
	panic("implement me")
}
