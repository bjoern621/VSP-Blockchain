package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/inv"

	"bjoernblessin.de/go-utils/util/logger"
)

func (b *Blockchain) Mempool(peerID common.PeerId) {
	logger.Infof("Mempool Message received from %v", peerID)

	// Get all transaction hashes from the mempool
	txHashes := b.mempool.GetAllTransactionHashes()

	if len(txHashes) == 0 {
		logger.Debugf("No transactions in mempool to announce to %v", peerID)
		return
	}

	// Create InvVectors for each transaction
	inventory := make([]*inv.InvVector, 0, len(txHashes))
	for _, hash := range txHashes {
		inventory = append(inventory, &inv.InvVector{
			Hash:    hash,
			InvType: inv.InvTypeMsgTx,
		})
	}

	// Send Inv message to the requesting peer
	b.blockchainMsgSender.SendInv(inventory, peerID)
	logger.Infof("Sent %d transaction hashes from mempool to %v", len(inventory), peerID)
}
