package konto

import (
	"s3b/vsp-blockchain/rest-api/internal/common"
	"s3b/vsp-blockchain/rest-api/vsgoin_node_adapter"
)

type KeyGenerator interface {
	GenerateKeyset() (common.Keyset, error)
	GetKeysetFromWIF(privateKeyWIF string) (common.Keyset, error)
}

type KeyGeneratorImpl struct {
	transactionAdapter vsgoin_node_adapter.TransactionAdapter
}

func NewKeyGeneratorImpl(transactionAdapter vsgoin_node_adapter.TransactionAdapter) *KeyGeneratorImpl {
	return &KeyGeneratorImpl{
		transactionAdapter: transactionAdapter,
	}
}

func (k KeyGeneratorImpl) GenerateKeyset() (common.Keyset, error) {
	return k.transactionAdapter.GenerateKeyset()
}

func (k KeyGeneratorImpl) GetKeysetFromWIF(privateKeyWIF string) (common.Keyset, error) {
	return k.transactionAdapter.GetKeysetFromWIF(privateKeyWIF)
}
