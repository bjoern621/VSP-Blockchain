package api

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/inv"
)

type BlockchainAPI interface {
	SendGetData(inventory []*inv.InvVector, peerId common.PeerId)
	BroadcastInvExclusionary(inventory []*inv.InvVector, peerId common.PeerId)
	BroadcastAddedBlocks(blockHashes []common.Hash, excludedPeerId common.PeerId)
	RequestMissingBlockHeaders(blockLocator block.BlockLocator, peerDd common.PeerId)
}
