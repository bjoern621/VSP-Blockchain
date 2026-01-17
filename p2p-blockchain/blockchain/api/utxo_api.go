package api

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/utxo"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
)

// UtxoLookupAPI provides a read-only view of UTXOs for external consumers.
// This is the API-level interface that should be used by other subsystems.
type UtxoLookupAPI interface {
	// GetUTXO retrieves an output by transaction ID and output index
	GetUTXO(txID transaction.TransactionID, outputIndex uint32) (transaction.Output, error)

	// GetUTXOsByPubKeyHash returns all UTXOs belonging to the given PubKeyHash.
	GetUTXOsByPubKeyHash(pubKeyHash transaction.PubKeyHash) ([]transaction.UTXO, error)
}

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
