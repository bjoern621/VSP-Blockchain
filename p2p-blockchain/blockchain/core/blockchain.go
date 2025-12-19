package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/validation"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api/blockchain/dto"

	"bjoernblessin.de/go-utils/util/logger"
)

type Blockchain struct {
	mempool              *Mempool
	sender               api.BlockchainService
	transactionValidator validation.ValidationService
}

func NewBlockchain(sender api.BlockchainService, transactionValidator validation.ValidationService) *Blockchain {
	return &Blockchain{
		mempool:              NewMempool(),
		sender:               sender,
		transactionValidator: transactionValidator,
	}
}

func (b *Blockchain) Inv(invMsg dto.InvMsgDTO, peerID common.PeerId) {
	invVectors := block.InvVectorsFromInvMsgDTO(invMsg)
	logger.Infof("Inv Message received: %v from %v", invVectors, peerID)

	unknownData := make([]block.InvVector, 0)

	for _, v := range invVectors {
		switch v.InvType {
		case block.InvTypeMsgBlock:
			panic("not implemented")
		case block.InvTypeMsgTx:
			if !b.mempool.IsKnownTransactionHash(v.Hash) || !b.IsTransactionKnown(v.Hash) {
				unknownData = append(unknownData, v)
			}
		case block.InvTypeMsgFilteredBlock:
			panic("not implemented")
		}
	}

	b.requestData(unknownData, peerID)
}

func (b *Blockchain) GetData(getDataMsg dto.GetDataMsgDTO, peerID common.PeerId) {
	invVectors := block.InvVectorFromGetDataDTO(getDataMsg)
	logger.Infof("GetData Message received: %v from %v", invVectors, peerID)
}

func (b *Blockchain) Block(blockMsg dto.BlockMsgDTO, peerID common.PeerId) {
	blockConverted := block.NewBlockFromDTO(blockMsg.Block)
	logger.Infof("Block Message received: %v from %v", blockConverted, peerID)
}

func (b *Blockchain) MerkleBlock(merkleBlockMsg dto.MerkleBlockMsgDTO, peerID common.PeerId) {
	merkleBlock := block.NewMerkleBlockFromDTO(merkleBlockMsg.MerkleBlock)
	logger.Infof("MerkleBlock Message received: %v from %v", merkleBlock, peerID)
}

func (b *Blockchain) Tx(txMsg dto.TxMsgDTO, peerID common.PeerId) {
	tx := transaction.NewTransactionFromDTO(txMsg.Transaction)

	isValid, err := b.transactionValidator.ValidateTransaction(&tx)
	if !isValid {
		logger.Errorf("Tx Message received from %v is invalid: %v", peerID, err)
		return
	}
	if b.mempool.IsKnownTransactionId(tx.Hash()) || b.IsTransactionKnownById(tx.Hash()) {
		logger.Infof("Tx Message already known: %v from %v", tx, peerID)
		return
	}

	isNew := b.mempool.AddTransaction(tx)
	if isNew {
		b.sender.BroadcastInv(dto.InvMsgDTO{
			Inventory: []dto.InvVectorDTO{
				block.FromTxToDtoInvVector(tx),
			},
		}, peerID)
	}

	logger.Infof("Tx Message received: %v from %v", tx, peerID)
}

func (b *Blockchain) GetHeaders(locator dto.BlockLocatorDTO, peerID common.PeerId) {
	blockLocator := block.NewBlockLocatorFromDTO(locator)
	logger.Infof("GetHeaders Message received: %v from %v", blockLocator, peerID)
}

func (b *Blockchain) Headers(blockHeaders dto.BlockHeadersDTO, peerID common.PeerId) {
	headers := block.NewBlockHeadersFromDTO(blockHeaders)
	logger.Infof("Headers Message received: %v from %v", headers, peerID)
}

func (b *Blockchain) SetFilter(setFilterRequest dto.SetFilterRequestDTO, peerID common.PeerId) {
	request := block.NewSetFilterRequestFromDTO(setFilterRequest)
	logger.Infof("setFilerRequest Message received: %v from %v", request, peerID)
}
func (b *Blockchain) Mempool(peerID common.PeerId) {
	logger.Infof("Mempool Message received from %v", peerID)
}

func (b *Blockchain) requestData(missingData []block.InvVector, id common.PeerId) {
	dtoInvVectors := block.ToDtoInvVectors(missingData)
	b.sender.SendGetData(dto.GetDataMsgDTO{Inventory: dtoInvVectors}, id)
}
