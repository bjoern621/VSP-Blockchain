package api

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
)

type BlockchainService interface {
	SendGetData(inventory []*block.InvVector, peerId common.PeerId)
	BroadcastInv(inventory []*block.InvVector, peerId common.PeerId)
}
