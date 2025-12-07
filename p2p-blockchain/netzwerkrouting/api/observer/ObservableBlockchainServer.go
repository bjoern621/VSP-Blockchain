package observer

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain"
)

// ObservableBlockchainServer defines the interface for an observable blockchain server.
// A new observer shall be attached to the server via the Attach() method and be removed via the Detach() method.
// The server shall also implement the corresponding methods to notify observers about changes.
type ObservableBlockchainServer interface {
	// Attach is called by the observer to attach itself to the server.
	Attach(o BlockchainObserver)
	// Detach is called by the observer to detach itself from the server.
	Detach(o BlockchainObserver)

	NotifyInv(invMsg *blockchain.InvMsg)
	NotifyGetData(getDataMsg *blockchain.GetDataMsg)
	NotifyBlock(blockMsg *blockchain.BlockMsg)
	NotifyMerkleBlock(merkleBlockMsg *blockchain.MerkleBlockMsg)
	NotifyTx(txMsg *blockchain.TxMsg)
	NotifyGetHeaders(blockLocator *blockchain.BlockLocator)
	NotifyHeaders(headers []*blockchain.BlockHeader)
	NotifySetFilterRequest(setFilterRequest *blockchain.SetFilterRequest)
	NotifyMempool()
}
