package observer

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api/blockchain/dto"
)

// BlockchainObserverAPI defines the interface for a blockchain observer.
// A blockchain observer is somebody interested in blockchain events.
// A blockchain observer shall be attached to a blockchain server via the ObservableBlockchainServerAPI.Attach() method
// and be removed via the ObservableBlockchainServerAPI.Detach() method on a valid ObservableBlockchainServerAPI.
// A blockchain observer shall implement the corresponding methods to handle blockchain events.
type BlockchainObserverAPI interface {
	Inv(invMsg dto.InvMsgDTO, peerID common.PeerId)
	GetData(getDataMsg dto.GetDataMsgDTO, peerID common.PeerId)
	Block(blockMsg dto.BlockMsgDTO, peerID common.PeerId)
	MerkleBlock(merkleBlockMsg dto.MerkleBlockMsgDTO, peerID common.PeerId)
	Tx(txMsg dto.TxMsgDTO, peerID common.PeerId)
	GetHeaders(locator dto.BlockLocatorDTO, peerID common.PeerId)
	Headers(headers dto.BlockHeadersDTO, peerID common.PeerId)
	SetFilter(setFilterRequest dto.SetFilterRequestDTO, peerID common.PeerId)
	Mempool(peerID common.PeerId)
}
