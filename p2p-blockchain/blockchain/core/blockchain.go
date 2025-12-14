package core

import (
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
	logger.Infof("Inv Message received: %v from %v", invMsg, peerID)
}

func (b *Blockchain) GetData(getDataMsg dto.GetDataMsgDTO, peerID peer.PeerID) {
	logger.Infof("GetData Message received: %v from %v", getDataMsg, peerID)
}

func (b *Blockchain) Block(blockMsg dto.BlockMsgDTO, peerID peer.PeerID) {
	logger.Infof("Block Message received: %v from %v", blockMsg, peerID)
}

func (b *Blockchain) MerkleBlock(merkleBlockMsg dto.MerkleBlockMsgDTO, peerID peer.PeerID) {
	logger.Infof("MerkleBlock Message received: %v from %v", merkleBlockMsg, peerID)
}

func (b *Blockchain) Tx(txMsg dto.TxMsgDTO, peerID peer.PeerID) {
	logger.Infof("Tx Message received: %v from %v", txMsg, peerID)
}

func (b *Blockchain) GetHeaders(locator dto.BlockLocatorDTO, peerID peer.PeerID) {
	logger.Infof("GetHeaders Message received: %v from %v", locator, peerID)
}

func (b *Blockchain) Headers(headers dto.BlockHeadersDTO, peerID peer.PeerID) {
	logger.Infof("Headers Message received: %v from %v", headers, peerID)
}

func (b *Blockchain) SetFilter(setFilterRequest dto.SetFilterRequestDTO, peerID peer.PeerID) {
	logger.Infof("setFilerRequest Message received: %v from %v", setFilterRequest, peerID)
}
func (b *Blockchain) Mempool(peerID peer.PeerID) {
	logger.Infof("Mempool Message received from %v", peerID)
}
