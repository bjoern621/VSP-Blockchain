package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/inv"

	"bjoernblessin.de/go-utils/util/logger"
)

func (b *Blockchain) Inv(inventory []*inv.InvVector, peerID common.PeerId) {
	if !b.CheckPeerIsConnected(peerID) {
		return
	}

	logger.Infof("[inv_handler] Inv Message received: %v from %v", inventory, peerID)

	unknownData := make([]*inv.InvVector, 0)

	for _, v := range inventory {
		switch v.InvType {
		case inv.InvTypeMsgBlock:
			if _, err := b.blockStore.GetBlockByHash(v.Hash); err != nil {
				unknownData = append(unknownData, v)
			}
		case inv.InvTypeMsgTx:
			if !b.mempool.IsKnownTransactionHash(v.Hash) || !b.IsTransactionKnown(v.Hash) {
				unknownData = append(unknownData, v)
			}
		case inv.InvTypeMsgFilteredBlock:
			panic("not implemented")
		}
	}

	logger.Infof("[inv_handler] Inv Message from %v has %d unknown items, Sending GetData...", peerID, len(unknownData))

	if len(unknownData) > 0 {
		b.blockchainMsgSender.SendGetData(unknownData, peerID)
	}
}
