package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/transaction"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api/blockchain/dto"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"

	"bjoernblessin.de/go-utils/util/logger"
)

type Blockchain struct {
}

func NewBlockchain() *Blockchain {
	return &Blockchain{}
}

func (b *Blockchain) Inv(invMsg dto.InvMsgDTO, peerID peer.PeerID) {
	invVectors := block.InvVectorsFromInvMsgDTO(invMsg)
	logger.Infof("Inv Message received: %v from %v", invVectors, peerID)
}

func (b *Blockchain) GetData(getDataMsg dto.GetDataMsgDTO, peerID peer.PeerID) {
	invVectors := block.InvVectorFromGetDataDTO(getDataMsg)
	logger.Infof("GetData Message received: %v from %v", invVectors, peerID)
}

func (b *Blockchain) Block(blockMsg dto.BlockMsgDTO, peerID peer.PeerID) {
	blockConverted := block.NewBlockFromDTO(blockMsg.Block)
	logger.Infof("Block Message received: %v from %v", blockConverted, peerID)
}

func (b *Blockchain) MerkleBlock(merkleBlockMsg dto.MerkleBlockMsgDTO, peerID peer.PeerID) {
	merkleBlock := block.NewMerkleBlockFromDTO(merkleBlockMsg.MerkleBlock)
	logger.Infof("MerkleBlock Message received: %v from %v", merkleBlock, peerID)
}

func (b *Blockchain) Tx(txMsg dto.TxMsgDTO, peerID peer.PeerID) {
	tx := transaction.NewTransactionFromDTO(txMsg.Transaction)
	logger.Infof("Tx Message received: %v from %v", tx, peerID)
}

func (b *Blockchain) GetHeaders(locator dto.BlockLocatorDTO, peerID peer.PeerID) {
	blockLocator := block.NewBlockLocatorFromDTO(locator)
	logger.Infof("GetHeaders Message received: %v from %v", blockLocator, peerID)
}

func (b *Blockchain) Headers(blockHeaders dto.BlockHeadersDTO, peerID peer.PeerID) {
	headers := block.NewBlockHeadersFromDTO(blockHeaders)
	logger.Infof("Headers Message received: %v from %v", headers, peerID)
}

func (b *Blockchain) SetFilter(setFilterRequest dto.SetFilterRequestDTO, peerID peer.PeerID) {
	request := block.NewSetFilterRequestFromDTO(setFilterRequest)
	logger.Infof("setFilerRequest Message received: %v from %v", request, peerID)
}
func (b *Blockchain) Mempool(peerID peer.PeerID) {
	logger.Infof("Mempool Message received from %v", peerID)
}
