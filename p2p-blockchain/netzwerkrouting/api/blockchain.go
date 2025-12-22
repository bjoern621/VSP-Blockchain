package api

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/inv"
)

type BlockchainAPI interface {
	SendGetData(inventory []*inv.InvVector, peerId common.PeerId)
	BroadcastInv(inventory []*inv.InvVector, peerId common.PeerId)
}
