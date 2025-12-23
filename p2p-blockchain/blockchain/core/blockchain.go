package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/validation"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/inv"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api"

	"bjoernblessin.de/go-utils/util/logger"
)

const invalidBlockMessageFormat = "Block Message received from %v is invalid: %v"

type Blockchain struct {
	mempool             *Mempool
	blockchainMsgSender api.BlockchainAPI

	transactionValidator validation.TransactionValidationAPI
	blockValidator       validation.BlockValidationAPI
}

func NewBlockchain(blockchainMsgSender api.BlockchainAPI, transactionValidator validation.TransactionValidationAPI, blockValidator validation.BlockValidationAPI) *Blockchain {
	return &Blockchain{
		mempool:              NewMempool(transactionValidator),
		blockchainMsgSender:  blockchainMsgSender,
		transactionValidator: transactionValidator,
		blockValidator:       blockValidator,
	}
}

func (b *Blockchain) Inv(inventory []*inv.InvVector, peerID common.PeerId) {
	logger.Infof("Inv Message received: %v from %v", inventory, peerID)

	unknownData := make([]*inv.InvVector, 0)

	for _, v := range inventory {
		switch v.InvType {
		case inv.InvTypeMsgBlock:
			//if !b.blockchain.IsBlockKnown(v.Hash) {
			//	  unknownData = append(unknownData, v)
			//}
			panic("not implemented")
		case inv.InvTypeMsgTx:
			if !b.mempool.IsKnownTransactionHash(v.Hash) || !b.IsTransactionKnown(v.Hash) {
				unknownData = append(unknownData, v)
			}
		case inv.InvTypeMsgFilteredBlock:
			panic("not implemented")
		}
	}

	b.requestData(unknownData, peerID)
}

func (b *Blockchain) GetData(inventory []*inv.InvVector, peerID common.PeerId) {
	logger.Infof("GetData Message received: %v from %v", inventory, peerID)
}

func (b *Blockchain) Block(receivedBlock block.Block, peerID common.PeerId) {

	defer recheckAllOrphanBlocks()

	if ok, err := b.blockValidator.SanityCheck(receivedBlock); !ok {
		logger.Warnf(invalidBlockMessageFormat, peerID, err)
		return
	}

	if ok, err := b.blockValidator.ValidateHeader(receivedBlock); !ok {
		logger.Warnf(invalidBlockMessageFormat, peerID, err)
		return
	}

	if ok, err := isOrphanBlock(receivedBlock); !ok {
		logger.Warnf("Blocks previous block is unknown: %v", err)
		addToOrphanPool(receivedBlock)
		requestMissingData(receivedBlock)
		return
	}

	if ok, err := b.blockValidator.FullValidation(receivedBlock); !ok {
		logger.Warnf(invalidBlockMessageFormat, peerID, err)
		return
	}

	if isPartOfMainChain(receivedBlock) {
		// entferne transactionen des blcoks aus dem mempool
		// add block to main chain -> Should this implicitly update the utxos?
	} else {
		// add block to second chain (should check for chain reorganisations)
	}

	//b.blockchainMsgSender.BroadcastInvExclusionary()

	logger.Infof("Block Message received: %v from %v", receivedBlock, peerID)
}

// If a block from orphan pool is accepted into any chain, there should be an inv message send
func recheckAllOrphanBlocks() {
	//TODO
}

func isPartOfMainChain(receivedBlock block.Block) bool {
	return false
}

func isOrphanBlock(block block.Block) (bool, error) {
	return true, nil
}

func addToOrphanPool(block block.Block) {

}

func requestMissingData(block block.Block) {
	// TODO: request missing data via a GetHeaders() and a getData Method
}

func (b *Blockchain) MerkleBlock(merkleBlock block.MerkleBlock, peerID common.PeerId) {
	logger.Infof("MerkleBlock Message received: %v from %v", merkleBlock, peerID)
}

// Tx processes a transaction message
// If the transaction is valid and not yet known, it is added to the mempool and broadcasted to other peers
func (b *Blockchain) Tx(tx transaction.Transaction, peerID common.PeerId) {

	isValid, err := b.transactionValidator.ValidateTransaction(&tx)
	if !isValid {
		logger.Errorf("Tx Message received from %v is invalid: %v", peerID, err)
		return
	}
	if b.mempool.IsKnownTransactionId(tx.TransactionId()) || b.IsTransactionKnownById(tx.TransactionId()) {
		logger.Infof("Tx Message already known: %v from %v", tx, peerID)
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

	logger.Infof("Tx Message received: %v from %v", tx, peerID)
}

func (b *Blockchain) GetHeaders(locator block.BlockLocator, peerID common.PeerId) {
	logger.Infof("GetHeaders Message received: %v from %v", locator, peerID)
}

func (b *Blockchain) Headers(blockHeaders []*block.BlockHeader, peerID common.PeerId) {
	logger.Infof("Headers Message received: %v from %v", blockHeaders, peerID)
}

func (b *Blockchain) SetFilter(setFilterRequest block.SetFilterRequest, peerID common.PeerId) {
	logger.Infof("setFilerRequest Message received: %v from %v", setFilterRequest, peerID)
}
func (b *Blockchain) Mempool(peerID common.PeerId) {
	logger.Infof("Mempool Message received from %v", peerID)
}

func (b *Blockchain) requestData(missingData []*inv.InvVector, id common.PeerId) {
	b.blockchainMsgSender.SendGetData(missingData, id)
}
