package observer

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"
)

// BlockchainObserverAPI defines the interface for a blockchain observer.
// A blockchain observer is somebody interested in blockchain events.
// A blockchain observer shall be attached to a blockchain server via the ObservableBlockchainServerAPI.Attach() method
// and be removed via the ObservableBlockchainServerAPI.Detach() method on a valid ObservableBlockchainServerAPI.
// A blockchain observer shall implement the corresponding methods to handle blockchain events.
type BlockchainObserverAPI interface {
	Inv(invMsg *blockchain.InvMsg, peerID peer.PeerID)
	GetData(getDataMsg *blockchain.GetDataMsg, peerID peer.PeerID)
	Block(blockMsg *blockchain.BlockMsg, peerID peer.PeerID)
	MerkleBlock(merkleBlockMsg *blockchain.MerkleBlockMsg, peerID peer.PeerID)
	Tx(txMsg *blockchain.TxMsg, peerID peer.PeerID)
	GetHeaders(locator *blockchain.BlockLocator, peerID peer.PeerID)
	Headers(headers []*blockchain.BlockHeader, peerID peer.PeerID)
	SetFilter(setFilterRequest *blockchain.SetFilterRequest, peerID peer.PeerID)
	Mempool(peerID peer.PeerID)
}
