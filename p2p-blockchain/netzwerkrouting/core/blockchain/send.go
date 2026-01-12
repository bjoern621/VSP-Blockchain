package blockchain

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/inv"
	"slices"
)

type BlockchainMsgSender interface {
	// SendGetData sends a getdata message to the given peer
	SendGetData(inventory []*inv.InvVector, peerId common.PeerId)

	// SendInv sends an inv message to the given peer
	SendInv(inventory []*inv.InvVector, peerId common.PeerId)

	// SendGetHeaders sends a GetHeaders message to the given peer
	SendGetHeaders(locator block.BlockLocator, peerId common.PeerId)
}

// SendGetData sends a getdata message to the given peer
func (b *BlockchainService) SendGetData(inventory []*inv.InvVector, peerId common.PeerId) {
	_, ok := b.peerRetriever.GetPeer(peerId)
	if !ok {
		panic("peer '" + peerId + "' not found")
	}
	b.blockchainMsgSender.SendGetData(inventory, peerId)
}

// BroadcastInvExclusionary propagates an inventory message to all outbound peers except the specified peer.
func (b *BlockchainService) BroadcastInvExclusionary(inventory []*inv.InvVector, excludedPeerId common.PeerId) {
	ids := b.peerRetriever.GetAllOutboundPeers()
	ownIndex := slices.Index(ids, excludedPeerId)
	ids = slices.Delete(ids, ownIndex, ownIndex+1)
	for _, id := range ids {
		b.blockchainMsgSender.SendInv(inventory, id)
	}
}

// BroadcastAddedBlocks broadcasts new block hashes to all outbound peers except the sender.
// Usually called when new blocks are added to the blockchain. This can happen, when a new block is mined locally
// or when new blocks are received from other peers and they are successfully validated and added to the side or main chain.
func (b *BlockchainService) BroadcastAddedBlocks(blockHashes []common.Hash, excludedPeerId common.PeerId) {
	if len(blockHashes) == 0 {
		return
	}

	invVectors := make([]*inv.InvVector, 0, len(blockHashes))
	for _, blockHash := range blockHashes {
		invVectors = append(invVectors, &inv.InvVector{
			Hash:    blockHash,
			InvType: inv.InvTypeMsgBlock,
		})
	}

	b.BroadcastInvExclusionary(invVectors, excludedPeerId)
}

// RequestMissingBlockHeaders sends a GetHeaders message to all outbound peers to request missing blocks.
// The locator is built using Fibonacci series to exponentially go back through the chain,
// allowing efficient synchronization even when chains have diverged significantly.
func (b *BlockchainService) RequestMissingBlockHeaders(blockLocator block.BlockLocator, peerId common.PeerId) {
	b.blockchainMsgSender.SendGetHeaders(blockLocator, peerId)
}
