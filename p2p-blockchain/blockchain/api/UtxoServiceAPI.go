package api

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/utxo"
)

type UtxoServiceAPI interface {
	utxo.UTXOService
}
