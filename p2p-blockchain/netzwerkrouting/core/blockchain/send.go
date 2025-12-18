package blockchain

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api/blockchain/dto"
	"slices"
)

type BlockchainMsgSender interface {
	SendGetData(dto dto.GetDataMsgDTO, peerId common.PeerId)
	SendInv(dto dto.InvMsgDTO, peerId common.PeerId)
}

func (b *BlockchainService) SendGetData(getDataMsg dto.GetDataMsgDTO, peerId common.PeerId) {
	_, ok := b.peerStore.GetPeer(peerId)
	if !ok {
		panic("peer '" + peerId + "' not found")
	}
	b.blockchainMsgSender.SendGetData(getDataMsg, peerId)
}

func (b *BlockchainService) BroadcastInv(invMsg dto.InvMsgDTO, peerId common.PeerId) {
	ids := b.peerStore.GetAllOutputPeers()
	ownIndex := slices.Index(ids, peerId)
	ids = slices.Delete(ids, ownIndex, ownIndex+1)
	for _, id := range ids {
		b.blockchainMsgSender.SendInv(invMsg, id)
	}
}
