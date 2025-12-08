package observer

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"
)

// ObservableBlockchainServerAPI defines the interface for an observable blockchain server.
// A new observer shall be attached to the server via the Attach() method and be removed via the Detach() method.
// The server shall also implement the corresponding methods to notify observers about changes.
type ObservableBlockchainServerAPI interface {
	// Attach is called by the observer to attach itself to the server.
	Attach(o BlockchainObserverAPI)
	// Detach is called by the observer to detach itself from the server.
	Detach(o BlockchainObserverAPI)

	NotifyInv(invMsg *blockchain.InvMsg, peerID peer.PeerID)
	NotifyGetData(getDataMsg *blockchain.GetDataMsg, peerID peer.PeerID)
	NotifyBlock(blockMsg *blockchain.BlockMsg, peerID peer.PeerID)
	NotifyMerkleBlock(merkleBlockMsg *blockchain.MerkleBlockMsg, peerID peer.PeerID)
	NotifyTx(txMsg *blockchain.TxMsg, peerID peer.PeerID)
	NotifyGetHeaders(blockLocator *blockchain.BlockLocator, peerID peer.PeerID)
	NotifyHeaders(headers []*blockchain.BlockHeader, peerID peer.PeerID)
	NotifySetFilterRequest(setFilterRequest *blockchain.SetFilterRequest, peerID peer.PeerID)
	NotifyMempool(peerID peer.PeerID)
}
