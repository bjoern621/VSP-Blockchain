package konto

import (
	"s3b/vsp-blockchain/rest-api/internal/common"
	"s3b/vsp-blockchain/rest-api/vsgoin_node_adapter"
)

type Kontostand interface {
	GenerateKeyset() common.Keyset
	GetKeysetFromWIF(privateKeyWIF string) (common.Keyset, error)
}

type KontostandImpl struct {
	transactionAdapter vsgoin_node_adapter.TransactionAdapter
}

func NewKontostand(transactionAdapter vsgoin_node_adapter.TransactionAdapter) *KontostandImpl {
	return &KontostandImpl{
		transactionAdapter: transactionAdapter,
	}
}

// not Implemented
func (k KontostandImpl) GenerateKeyset() common.Keyset {
	return common.Keyset{
		PrivateKeyWif: "private Key Wif",
		VSAddress:     "VS Address",
	}
}

// not Implemented
func (k KontostandImpl) GetKeysetFromWIF(privateKeyWIF string) (common.Keyset, error) {
	return common.Keyset{
		PrivateKeyWif: privateKeyWIF,
		VSAddress:     "VS Address",
	}, nil //TODO: zwischen internal und private key error unterscheiden
}
