package observer

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/inv"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
)

// BlockchainObserverAPI defines the interface for a blockchain observer.
// A blockchain observer is somebody interested in blockchain events.
// A blockchain observer shall be attached to a blockchain server via the ObservableBlockchainServerAPI.Attach() method
// and be removed via the ObservableBlockchainServerAPI.Detach() method on a valid ObservableBlockchainServerAPI.
// A blockchain observer shall implement the corresponding methods to handle blockchain events.
type BlockchainObserverAPI interface {
	Inv(inventory []*inv.InvVector, peerID common.PeerId)
	GetData(inventory []*inv.InvVector, peerID common.PeerId)
	Block(block block.Block, peerID common.PeerId)
	MerkleBlock(merkleBlock block.MerkleBlock, peerID common.PeerId)
	Tx(tx transaction.Transaction, peerID common.PeerId)
	GetHeaders(locator block.BlockLocator, peerID common.PeerId)
	Headers(headers []*block.BlockHeader, peerID common.PeerId)
	SetFilter(setFilterRequest block.SetFilterRequest, peerID common.PeerId)
	Mempool(peerID common.PeerId)
}
