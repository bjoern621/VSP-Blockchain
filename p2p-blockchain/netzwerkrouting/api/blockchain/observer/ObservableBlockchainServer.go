package observer

import (
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api/blockchain/dto"
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

	NotifyInv(invMsg dto.InvMsgDTO, peerID peer.PeerID)
	NotifyGetData(getDataMsg dto.GetDataMsgDTO, peerID peer.PeerID)
	NotifyBlock(blockMsg dto.BlockMsgDTO, peerID peer.PeerID)
	NotifyMerkleBlock(merkleBlockMsg dto.MerkleBlockMsgDTO, peerID peer.PeerID)
	NotifyTx(txMsg dto.TxMsgDTO, peerID peer.PeerID)
	NotifyGetHeaders(blockLocator dto.BlockLocatorDTO, peerID peer.PeerID)
	NotifyHeaders(headers dto.BlockHeadersDTO, peerID peer.PeerID)
	NotifySetFilterRequest(setFilterRequest dto.SetFilterRequestDTO, peerID peer.PeerID)
	NotifyMempool(peerID peer.PeerID)
}
