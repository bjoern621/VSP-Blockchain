package blockchain

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"slices"
)

type BlockchainMsgSender interface {
	SendGetData(inventory []*block.InvVector, peerId common.PeerId)
	SendInv(inventory []*block.InvVector, peerId common.PeerId)
}

func (b *BlockchainService) SendGetData(inventory []*block.InvVector, peerId common.PeerId) {
	_, ok := b.peerStore.GetPeer(peerId)
	if !ok {
		panic("peer '" + peerId + "' not found")
	}
	b.blockchainMsgSender.SendGetData(inventory, peerId)
}

func (b *BlockchainService) BroadcastInv(inventory []*block.InvVector, peerId common.PeerId) {
	ids := b.peerStore.GetAllOutputPeers()
	ownIndex := slices.Index(ids, peerId)
	ids = slices.Delete(ids, ownIndex, ownIndex+1)
	for _, id := range ids {
		b.blockchainMsgSender.SendInv(inventory, id)
	}
}
