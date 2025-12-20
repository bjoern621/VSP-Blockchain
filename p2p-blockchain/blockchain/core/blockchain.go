package core

import (
	api2 "s3b/vsp-blockchain/p2p-blockchain/blockchain/api"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/validation"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api"

	"bjoernblessin.de/go-utils/util/logger"
)

type Blockchain struct {
	mempool              *Mempool
	sender               api.BlockchainService
	transactionValidator validation.ValidationService
	//TODO: use from https://github.com/bjoern621/VSP-Blockchain/tree/166-task-blockchain-und-header-datenstruktur-erstellen
	blockStore api2.BlockStore
}

func NewBlockchain(sender api.BlockchainService, transactionValidator validation.ValidationService) *Blockchain {
	return &Blockchain{
		mempool:              NewMempool(),
		sender:               sender,
		transactionValidator: transactionValidator,
	}
}

func (b *Blockchain) Inv(inventory []*block.InvVector, peerID common.PeerId) {
	logger.Infof("Inv Message received: %v from %v", inventory, peerID)

	dataToRequest := make([]*block.InvVector, 0)

	for _, v := range inventory {
		switch v.InvType {
		case block.InvTypeMsgBlock:
			if !b.blockStore.IsKnownBlockHash(v.Hash) {
				dataToRequest = append(dataToRequest, v)
			}
		case block.InvTypeMsgTx:
			if !b.mempool.IsKnownTransactionHash(v.Hash) || !b.IsTransactionKnown(v.Hash) {
				dataToRequest = append(dataToRequest, v)
			}
		case block.InvTypeMsgFilteredBlock:
			panic("not implemented")
		}
	}

	b.requestData(dataToRequest, peerID)
}

func (b *Blockchain) GetData(inventory []*block.InvVector, peerID common.PeerId) {
	logger.Infof("GetData Message received: %v from %v", inventory, peerID)
}

func (b *Blockchain) Block(block block.Block, peerID common.PeerId) {
	//if !b.blockStore.IsKnownBlockHash(block.Header)

	logger.Infof("Block Message received: %v from %v", block, peerID)
}

func (b *Blockchain) MerkleBlock(merkleBlock block.MerkleBlock, peerID common.PeerId) {
	logger.Infof("MerkleBlock Message received: %v from %v", merkleBlock, peerID)
}

func (b *Blockchain) Tx(tx transaction.Transaction, peerID common.PeerId) {
	logger.Infof("Tx Message received: %v from %v", tx, peerID)

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
		invVectors := make([]*block.InvVector, 0)
		invVector := block.InvVector{
			Hash:    tx.Hash(),
			InvType: block.InvTypeMsgTx,
		}
		invVectors = append(invVectors, &invVector)
		b.sender.BroadcastInv(invVectors, peerID)
	}
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

func (b *Blockchain) requestData(missingData []*block.InvVector, id common.PeerId) {
	b.sender.SendGetData(missingData, id)
}
