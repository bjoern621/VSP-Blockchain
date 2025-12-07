package observer

import "s3b/vsp-blockchain/p2p-blockchain/blockchain"

// BlockchainObserver defines the interface for a blockchain observer.
// A blockchain observer is somebody interested in blockchain events.
// A blockchain observer shall be attached to a blockchain server via the ObservableBlockchainServer.Attach() method
// and be removed via the ObservableBlockchainServer.Detach() method on a valid ObservableBlockchainServer.
// A blockchain observer shall implement the corresponding methods to handle blockchain events.
type BlockchainObserver interface {
	Inv(invMsg *blockchain.InvMsg)
	GetData(getDataMsg *blockchain.GetDataMsg)
	Block(blockMsg *blockchain.BlockMsg)
	MerkleBlock(merkleBlockMsg *blockchain.MerkleBlockMsg)
	Tx(txMsg *blockchain.TxMsg)
	GetHeaders(locator *blockchain.BlockLocator)
	Headers(headers []*blockchain.BlockHeader)
	SetFilter(setFilterRequest *blockchain.SetFilterRequest)
	Mempool()
}
