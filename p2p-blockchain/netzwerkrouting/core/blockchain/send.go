package blockchain

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/inv"
	"slices"
)

type BlockchainMsgSender interface {
	SendGetData(inventory []*inv.InvVector, peerId common.PeerId)
	SendInv(inventory []*inv.InvVector, peerId common.PeerId)
	SendGetHeaders(locator block.BlockLocator, peerId common.PeerId)
}

// SendGetData sends a getdata message to the given peer
func (b *BlockchainService) SendGetData(inventory []*inv.InvVector, peerId common.PeerId) {
	_, ok := b.peerStore.GetPeer(peerId)
	if !ok {
		panic("peer '" + peerId + "' not found")
	}
	b.blockchainMsgSender.SendGetData(inventory, peerId)
}

// BroadcastInvExclusionary propagates an inventory message to all outbound peers except the specified peer.
func (b *BlockchainService) BroadcastInvExclusionary(inventory []*inv.InvVector, excludedPeerId common.PeerId) {
	ids := b.peerStore.GetAllOutboundPeers()
	ownIndex := slices.Index(ids, excludedPeerId)
	ids = slices.Delete(ids, ownIndex, ownIndex+1)
	for _, id := range ids {
		b.blockchainMsgSender.SendInv(inventory, id)
	}
}

// BroadcastAddedBlocks broadcasts new block hashes to all outbound peers except the sender.
func (b *BlockchainService) BroadcastAddedBlocks(blockHashes []common.Hash, excludedPeerId common.PeerId) {
	if len(blockHashes) == 0 {
		return
	}

	// Convert block hashes to inventory vectors
	invVectors := make([]*inv.InvVector, 0, len(blockHashes))
	for _, blockHash := range blockHashes {
		invVectors = append(invVectors, &inv.InvVector{
			Hash:    blockHash,
			InvType: inv.InvTypeMsgBlock,
		})
	}

	// Broadcast using the existing BroadcastInvExclusionary method
	b.BroadcastInvExclusionary(invVectors, excludedPeerId)
}

// RequestMissingBlockHeaders sends a GetHeaders message to all outbound peers to request missing blocks.
// The locator is built from the orphan block's parent hash.
func (b *BlockchainService) RequestMissingBlockHeaders(orphanParentHash common.Hash) {
	// Create a block locator with the parent hash
	// This tells peers: "I have this hash, send me headers after it"
	locator := block.BlockLocator{
		BlockLocatorHashes: []common.Hash{orphanParentHash},
		StopHash:           common.Hash{}, // Empty stop hash means don't stop
	}

	// Send to all outbound peers
	allPeers := b.peerStore.GetAllOutboundPeers()
	for _, peerID := range allPeers {
		b.blockchainMsgSender.SendGetHeaders(locator, peerID)
	}
}
