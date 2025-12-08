package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api/observer"
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

func (b *Blockchain) Inv(invMsg *blockchain.InvMsg) {
	logger.Infof("Inv Message received: %v", invMsg)
}

func (b *Blockchain) GetData(getDataMsg *blockchain.GetDataMsg) {
	logger.Infof("GetData Message received: %v", getDataMsg)
}

func (b *Blockchain) Block(blockMsg *blockchain.BlockMsg) {
	logger.Infof("Block Message received: %v", blockMsg)
}

func (b *Blockchain) MerkleBlock(merkleBlockMsg *blockchain.MerkleBlockMsg) {
	logger.Infof("MerkleBlock Message received: %v", merkleBlockMsg)
}

func (b *Blockchain) Tx(txMsg *blockchain.TxMsg) {
	logger.Infof("Tx Message received: %v", txMsg)
}

func (b *Blockchain) GetHeaders(locator *blockchain.BlockLocator) {
	logger.Infof("GetHeaders Message received: %v", locator)
}

func (b *Blockchain) Headers(headers []*blockchain.BlockHeader) {
	logger.Infof("Headers Message received: %v", headers)
}

func (b *Blockchain) SetFilter(setFilterRequest *blockchain.SetFilterRequest) {
	logger.Infof("setFilerRequest Message received: %v", setFilterRequest)
}
func (b *Blockchain) Mempool() {
	logger.Infof("Mempool Message received")
}
