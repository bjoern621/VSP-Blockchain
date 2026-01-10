package api

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
)

type BlockchainAPI interface {
	Block(receivedBlock block.Block, peerID common.PeerId)
}
