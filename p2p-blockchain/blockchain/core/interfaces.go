package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/blockchain"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/utxo"
)

type BlockStoreAPI interface {
	blockchain.BlockStoreAPI
}

type UtxoServiceAPI interface {
	utxo.UTXOService
}
