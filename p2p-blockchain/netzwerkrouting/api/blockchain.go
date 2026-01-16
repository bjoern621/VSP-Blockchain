package api

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/inv"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
)

type BlockchainAPI interface {
	// SendGetData sends a getdata message to the given peer
	SendGetData(inventory []*inv.InvVector, peerId common.PeerId)

	// SendInv sends an inv message to the given peer
	SendInv(inventory []*inv.InvVector, peerId common.PeerId)

	// BroadcastInvExclusionary propagates an inventory message to all outbound peers except the specified peer.
	BroadcastInvExclusionary(inventory []*inv.InvVector, peerId common.PeerId)

	// BroadcastAddedBlocks broadcasts new block hashes to all outbound peers except the sender.
	// Usually called when new blocks are added to the blockchain. This can happen, when a new block is mined locally
	// or when new blocks are received from other peers and they are successfully validated and added to the side or main chain.
	BroadcastAddedBlocks(blockHashes []common.Hash, excludedPeerId common.PeerId)

	// RequestMissingBlockHeaders sends a GetHeaders message to all outbound peers to request missing blocks.
	// The locator is built using Fibonacci series to exponentially go back through the chain,
	// allowing efficient synchronization even when chains have diverged significantly.
	RequestMissingBlockHeaders(blockLocator block.BlockLocator, peerDd common.PeerId)

	// SendHeaders sends a Headers message to the given peer
	SendHeaders(headers []*block.BlockHeader, peerId common.PeerId)
}

// FullInventoryInformationMsgSenderAPI defines methods to send full inventory information messages.
// Used by the blockchain to send full block and transaction data to peers in response to getdata requests.
// This interface is separate from BlockchainAPI to adhere to the single responsibility principle and avoid bloating the interface.
// Implemented by the infrastructure layer to handle serialization and network transmission of full inventory messages.
type FullInventoryInformationMsgSenderAPI interface {
	// SendBlock sends a Block message to the given peer
	SendBlock(block block.Block, peerId common.PeerId)

	// SendTx sends a Tx message to the given peer
	SendTx(tx transaction.Transaction, peerId common.PeerId)
}
