package blockchain

import (
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/data/peer"
)

type BlockchainService struct {
	blockchainMsgSender BlockchainMsgSender
	peerRetriever       peer.PeerRetriever
}

func NewBlockchainService(blockchainMsgSender BlockchainMsgSender, peerRetriever peer.PeerRetriever) *BlockchainService {
	return &BlockchainService{
		blockchainMsgSender: blockchainMsgSender,
		peerRetriever:       peerRetriever,
	}
}
