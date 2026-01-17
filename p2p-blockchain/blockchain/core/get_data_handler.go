package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/inv"

	"bjoernblessin.de/go-utils/util/assert"
	"bjoernblessin.de/go-utils/util/logger"
)

func (b *Blockchain) GetData(inventory []*inv.InvVector, peerID common.PeerId) {
	logger.Infof("GetData Message received: %v from %v", inventory, peerID)

	for _, invVector := range inventory {
		switch invVector.InvType {
		case inv.InvTypeMsgBlock:
			b.handleBlockRequest(invVector.Hash, peerID)
		case inv.InvTypeMsgTx:
			b.handleTransactionRequest(invVector.Hash, peerID)
		case inv.InvTypeMsgFilteredBlock:
			assert.Never("not supported")
		}
	}
}

func (b *Blockchain) handleBlockRequest(blockHash common.Hash, peerID common.PeerId) {
	block, err := b.blockStore.GetBlockByHash(blockHash)
	if err != nil {
		logger.Warnf("Requested block %v not found for GetData from peer %v", blockHash, peerID)
		return
	}

	go b.fullInventoryMsgSender.SendBlock(block, peerID)
}

func (b *Blockchain) handleTransactionRequest(txHash common.Hash, peerID common.PeerId) {
	transaction, found := b.mempool.GetTransactionByHash(txHash)
	if !found {
		logger.Warnf("Requested transaction %v not found for GetData from peer %v", txHash, peerID)
		return
	}

	go b.fullInventoryMsgSender.SendTx(transaction, peerID)
}
