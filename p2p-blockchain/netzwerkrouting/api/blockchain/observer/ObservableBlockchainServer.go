package observer

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
)

// ObservableBlockchainServerAPI defines the interface for an observable blockchain server.
// A new observer shall be attached to the server via the Attach() method and be removed via the Detach() method.
// The server shall also implement the corresponding methods to notify observers about changes.
type ObservableBlockchainServerAPI interface {
	// Attach is called by the observer to attach itself to the server.
	Attach(o BlockchainObserverAPI)
	// Detach is called by the observer to detach itself from the server.
	Detach(o BlockchainObserverAPI)

	NotifyInv(inventory []*block.InvVector, peerID common.PeerId)
	NotifyGetData(inventory []*block.InvVector, peerID common.PeerId)
	NotifyBlock(block block.Block, peerID common.PeerId)
	NotifyMerkleBlock(merkleBlock block.MerkleBlock, peerID common.PeerId)
	NotifyTx(tx transaction.Transaction, peerID common.PeerId)
	NotifyGetHeaders(locator block.BlockLocator, peerID common.PeerId)
	NotifyHeaders(headers []*block.BlockHeader, peerID common.PeerId)
	NotifySetFilter(setFilterRequest block.SetFilterRequest, peerID common.PeerId)
	NotifyMempool(peerID common.PeerId)
}
