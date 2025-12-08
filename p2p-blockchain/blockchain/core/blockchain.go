package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api/observer"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/infrastructure/middleware/grpc"

	"bjoernblessin.de/go-utils/util/logger"
)

type Blockchain struct {
	server *grpc.Server
}

var _ observer.BlockchainObserverAPI = (*Blockchain)(nil)

func NewBlockchain(server *grpc.Server) *Blockchain {
	chain := &Blockchain{
		server: server,
	}

	server.Attach(chain)

	return chain
}

func (b *Blockchain) Inv(invMsg *blockchain.InvMsg, peerID peer.PeerID) {
	logger.Infof("Inv Message received: %v from %v", invMsg, peerID)
}

func (b *Blockchain) GetData(getDataMsg *blockchain.GetDataMsg, peerID peer.PeerID) {
	logger.Infof("GetData Message received: %v from %v", getDataMsg, peerID)
}

func (b *Blockchain) Block(blockMsg *blockchain.BlockMsg, peerID peer.PeerID) {
	logger.Infof("Block Message received: %v from %v", blockMsg, peerID)
}

func (b *Blockchain) MerkleBlock(merkleBlockMsg *blockchain.MerkleBlockMsg, peerID peer.PeerID) {
	logger.Infof("MerkleBlock Message received: %v from %v", merkleBlockMsg, peerID)
}

func (b *Blockchain) Tx(txMsg *blockchain.TxMsg, peerID peer.PeerID) {
	logger.Infof("Tx Message received: %v from %v", txMsg, peerID)
}

func (b *Blockchain) GetHeaders(locator *blockchain.BlockLocator, peerID peer.PeerID) {
	logger.Infof("GetHeaders Message received: %v from %v", locator, peerID)
}

func (b *Blockchain) Headers(headers []*blockchain.BlockHeader, peerID peer.PeerID) {
	logger.Infof("Headers Message received: %v from %v", headers, peerID)
}

func (b *Blockchain) SetFilter(setFilterRequest *blockchain.SetFilterRequest, peerID peer.PeerID) {
	logger.Infof("setFilerRequest Message received: %v from %v", setFilterRequest, peerID)
}
func (b *Blockchain) Mempool(peerID peer.PeerID) {
	logger.Infof("Mempool Message received from %v", peerID)
}
