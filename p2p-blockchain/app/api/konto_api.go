package api

import (
	"s3b/vsp-blockchain/p2p-blockchain/app/data"
	blockapi "s3b/vsp-blockchain/p2p-blockchain/blockchain/api"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
)

// KontoAPI provides the interface for querying konto (account) information.
type KontoAPI interface {
	// GetAssets returns the assets (UTXO values) for a given V$Address.
	GetAssets(vsAddress string) data.AssetsResult
}

// KontoAPIImpl implements KontoAPI using the UTXO API.
type KontoAPIImpl struct {
	utxoAPI    *blockapi.UtxoAPI
	keyDecoder VSAddressDecoder
}

// VSAddressDecoder decodes V$Address strings to public key hashes.
type VSAddressDecoder interface {
	Base58CheckToBytes(input string) ([]byte, byte, error)
}

// NewKontoAPIImpl creates a new KontoAPIImpl with the given dependencies.
func NewKontoAPIImpl(utxoAPI *blockapi.UtxoAPI, keyDecoder VSAddressDecoder) *KontoAPIImpl {
	return &KontoAPIImpl{
		utxoAPI:    utxoAPI,
		keyDecoder: keyDecoder,
	}
}

// GetAssets implements KontoAPI.GetAssets.
func (api *KontoAPIImpl) GetAssets(vsAddress string) data.AssetsResult {
	// Decode the V$Address to get the public key hash
	pubKeyHashBytes, version, err := api.keyDecoder.Base58CheckToBytes(vsAddress)
	result, ok := api.validatePubKeyHash(err, version, pubKeyHashBytes)
	if !ok {
		return result
	}

	var pubKeyHash [20]byte
	copy(pubKeyHash[:], pubKeyHashBytes)

	utxos, err := api.utxoAPI.GetUTXOsByPubKeyHash(pubKeyHash)
	if err != nil {
		return data.AssetsResult{
			Success:      false,
			ErrorMessage: "failed to query UTXOs: " + err.Error(),
		}
	}

	return api.handleSuccess(utxos)
}

func (api *KontoAPIImpl) handleSuccess(utxos []transaction.UTXO) data.AssetsResult {
	assets := make([]data.Asset, 0, len(utxos))
	for _, utxo := range utxos {
		assets = append(assets, data.Asset{
			Value: utxo.Output.Value,
		})
	}

	return data.AssetsResult{
		Success: true,
		Assets:  assets,
	}
}

func (api *KontoAPIImpl) validatePubKeyHash(err error, version byte, pubKeyHashBytes []byte) (data.AssetsResult, bool) {
	if err != nil {
		return data.AssetsResult{
			Success:      false,
			ErrorMessage: "invalid V$Address format: " + err.Error(),
		}, false
	}

	// V$Address uses version 0x00
	if version != 0x00 {
		return data.AssetsResult{
			Success:      false,
			ErrorMessage: "invalid V$Address version byte",
		}, false
	}

	// Convert to fixed-size array (20 bytes for public key hash)
	if len(pubKeyHashBytes) != 20 {
		return data.AssetsResult{
			Success:      false,
			ErrorMessage: "invalid public key hash length",
		}, false
	}
	return data.AssetsResult{}, true
}
