package blockchain

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/blockchain"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"
)

type BlockchainService struct {
	blockchainMsgSender BlockchainMsgSender
	peerStore           *peer.PeerStore
	blockStore          *blockchain.BlockStore
}

func NewBlockchainService(blockchainMsgSender BlockchainMsgSender, peerStore *peer.PeerStore) *BlockchainService {
	return &BlockchainService{
		blockchainMsgSender: blockchainMsgSender,
		peerStore:           peerStore,
	}
}
