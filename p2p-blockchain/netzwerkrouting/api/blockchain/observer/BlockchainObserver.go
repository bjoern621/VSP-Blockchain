package observer

import (
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api/blockchain/dto"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"
)

// BlockchainObserverAPI defines the interface for a blockchain observer.
// A blockchain observer is somebody interested in blockchain events.
// A blockchain observer shall be attached to a blockchain server via the ObservableBlockchainServerAPI.Attach() method
// and be removed via the ObservableBlockchainServerAPI.Detach() method on a valid ObservableBlockchainServerAPI.
// A blockchain observer shall implement the corresponding methods to handle blockchain events.
type BlockchainObserverAPI interface {
	Inv(invMsg dto.InvMsgDTO, peerID peer.PeerID)
	GetData(getDataMsg dto.GetDataMsgDTO, peerID peer.PeerID)
	Block(blockMsg dto.BlockMsgDTO, peerID peer.PeerID)
	MerkleBlock(merkleBlockMsg dto.MerkleBlockMsgDTO, peerID peer.PeerID)
	Tx(txMsg dto.TxMsgDTO, peerID peer.PeerID)
	GetHeaders(locator dto.BlockLocatorDTO, peerID peer.PeerID)
	Headers(headers dto.BlockHeadersDTO, peerID peer.PeerID)
	SetFilter(setFilterRequest dto.SetFilterRequestDTO, peerID peer.PeerID)
	Mempool(peerID peer.PeerID)
}
