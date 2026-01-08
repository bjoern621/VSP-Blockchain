package blockchain

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/inv"
	"slices"
)

type BlockchainMsgSender interface {
	SendGetData(inventory []*inv.InvVector, peerId common.PeerId)
	SendInv(inventory []*inv.InvVector, peerId common.PeerId)
}

// SendGetData sends a getdata message to the given peer
func (b *BlockchainService) SendGetData(inventory []*inv.InvVector, peerId common.PeerId) {
	_, ok := b.peerRetriever.GetPeer(peerId)
	if !ok {
		panic("peer '" + peerId + "' not found")
	}
	b.blockchainMsgSender.SendGetData(inventory, peerId)
}

// BroadcastInvExclusionary propagates an inventory message to all outbound peers except the specified peer.
func (b *BlockchainService) BroadcastInvExclusionary(inventory []*inv.InvVector, excludedPeerId common.PeerId) {
	ids := b.peerRetriever.GetAllOutboundPeers()
	ownIndex := slices.Index(ids, excludedPeerId)
	ids = slices.Delete(ids, ownIndex, ownIndex+1)
	for _, id := range ids {
		b.blockchainMsgSender.SendInv(inventory, id)
	}
}
