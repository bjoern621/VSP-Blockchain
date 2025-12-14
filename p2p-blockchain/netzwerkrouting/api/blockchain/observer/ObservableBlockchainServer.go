package observer

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api/blockchain/dto"
)

// ObservableBlockchainServerAPI defines the interface for an observable blockchain server.
// A new observer shall be attached to the server via the Attach() method and be removed via the Detach() method.
// The server shall also implement the corresponding methods to notify observers about changes.
type ObservableBlockchainServerAPI interface {
	// Attach is called by the observer to attach itself to the server.
	Attach(o BlockchainObserverAPI)
	// Detach is called by the observer to detach itself from the server.
	Detach(o BlockchainObserverAPI)

	NotifyInv(invMsg dto.InvMsgDTO, peerID common.PeerId)
	NotifyGetData(getDataMsg dto.GetDataMsgDTO, peerID common.PeerId)
	NotifyBlock(blockMsg dto.BlockMsgDTO, peerID common.PeerId)
	NotifyMerkleBlock(merkleBlockMsg dto.MerkleBlockMsgDTO, peerID common.PeerId)
	NotifyTx(txMsg dto.TxMsgDTO, peerID common.PeerId)
	NotifyGetHeaders(blockLocator dto.BlockLocatorDTO, peerID common.PeerId)
	NotifyHeaders(headers dto.BlockHeadersDTO, peerID common.PeerId)
	NotifySetFilterRequest(setFilterRequest dto.SetFilterRequestDTO, peerID common.PeerId)
	NotifyMempool(peerID common.PeerId)
}
