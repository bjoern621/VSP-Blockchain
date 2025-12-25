package api

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/utxo"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
)

type UtxoAPI struct {
	utxo.LookupService
}

func NewUtxoAPI(lookupService utxo.LookupService) *UtxoAPI {
	return &UtxoAPI{
		LookupService: lookupService,
	}
}

func (u *UtxoAPI) GetUTXOsByPubKeyHash(pubKeyHash transaction.PubKeyHash) ([]transaction.UTXO, error) {
	return u.LookupService.GetUTXOsByPubKeyHash(pubKeyHash)
}
