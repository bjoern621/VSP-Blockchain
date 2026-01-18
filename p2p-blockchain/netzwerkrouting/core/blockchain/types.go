package blockchain

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
)

type BlockchainService struct {
	blockchainMsgSender BlockchainMsgSender
	peerRetriever       peerRetriever
}

func NewBlockchainService(blockchainMsgSender BlockchainMsgSender, peerRetriever peerRetriever) *BlockchainService {
	return &BlockchainService{
		blockchainMsgSender: blockchainMsgSender,
		peerRetriever:       peerRetriever,
	}
}

// peerRetriever is an interface for retrieving peers.
// It is implemented by peer.PeerStore.
type peerRetriever interface {
	GetPeer(id common.PeerId) (*common.Peer, bool)
	GetAllConnectedPeers() []common.PeerId
}
