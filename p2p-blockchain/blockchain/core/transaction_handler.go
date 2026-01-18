package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/inv"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"

	"bjoernblessin.de/go-utils/util/logger"
)

// Tx processes a transaction message
// If the transaction is valid and not yet known, it is added to the mempool and broadcasted to other peers
func (b *Blockchain) Tx(tx transaction.Transaction, peerID common.PeerId) {
	logger.Infof("[transaction_handler] Tx Message received: %v from %v", tx, peerID)

	mainChainTip := b.blockStore.GetMainChainTip()
	mainChainTipHash := mainChainTip.Hash()
	isValid, err := b.transactionValidator.ValidateTransaction(tx, mainChainTipHash)
	if !isValid {
		logger.Errorf("[transaction_handler] Tx Message received from %v is invalid: %v", peerID, err)
		return
	}
	if b.mempool.IsKnownTransactionId(tx.TransactionId()) || b.IsTransactionKnownById(tx.TransactionId()) {
		logger.Infof("[transaction_handler] Tx Message already known: %v from %v", tx, peerID)
		return
	}

	isNew := b.mempool.AddTransaction(tx)
	if isNew {
		invVectors := make([]*inv.InvVector, 0)
		invVector := inv.InvVector{
			Hash:    tx.Hash(),
			InvType: inv.InvTypeMsgTx,
		}
		invVectors = append(invVectors, &invVector)
		b.blockchainMsgSender.BroadcastInvExclusionary(invVectors, peerID)
	}
}
