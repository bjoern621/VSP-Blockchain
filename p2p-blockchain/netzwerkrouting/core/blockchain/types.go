package blockchain

import (
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/data/peer"
)

type BlockchainService struct {
	blockchainMsgSender BlockchainMsgSender
	peerStore           *peer.PeerStore
}

func NewBlockchainService(blockchainMsgSender BlockchainMsgSender, peerStore *peer.PeerStore) *BlockchainService {
	return &BlockchainService{
		blockchainMsgSender: blockchainMsgSender,
		peerStore:           peerStore,
	}
}
