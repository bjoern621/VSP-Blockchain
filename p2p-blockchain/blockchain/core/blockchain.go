package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/transaction"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api/blockchain/dto"

	"bjoernblessin.de/go-utils/util/logger"
)

type Blockchain struct {
}

func NewBlockchain() *Blockchain {
	return &Blockchain{}
}

func (b *Blockchain) Inv(invMsg dto.InvMsgDTO, peerID common.PeerId) {
	invVectors := block.InvVectorsFromInvMsgDTO(invMsg)
	logger.Infof("Inv Message received: %v from %v", invVectors, peerID)
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
