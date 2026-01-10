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
// The locator is built using Fibonacci series to exponentially go back through the chain,
// allowing efficient synchronization even when chains have diverged significantly.
func (b *BlockchainService) RequestMissingBlockHeaders(orphanParentHash common.Hash, peerId common.PeerId) {
	currentHeight := b.blockStore.GetCurrentHeight()
	locatorHashes := b.buildBlockLocator(currentHeight)

	// Prepend the orphan parent hash at the beginning (most recent hash)
	locatorHashes = append([]common.Hash{orphanParentHash}, locatorHashes...)

	locator := block.BlockLocator{
		BlockLocatorHashes: locatorHashes,
		StopHash:           common.Hash{}, // Empty stop hash means don't stop until we find common ancestor
	}

	b.blockchainMsgSender.SendGetHeaders(locator, peerId)
}

// buildBlockLocator creates a block locator using Fibonacci series to sample the chain.
// Returns hashes starting from newer blocks (closer to tip) to older blocks (closer to genesis).
func (b *BlockchainService) buildBlockLocator(tipHeight uint64) []common.Hash {
	locatorHashes := make([]common.Hash, 0)

	fib1, fib2 := uint64(1), uint64(2)
	offset := uint64(0)

	for offset <= tipHeight {
		height := tipHeight - offset

		blocksAtHeight := b.blockStore.GetBlocksByHeight(height)

		for _, blk := range blocksAtHeight {
			if b.blockStore.IsPartOfMainChain(blk) {
				locatorHashes = append(locatorHashes, blk.Hash())
				break
			}
		}

		offset += fib1
		fib1, fib2 = fib2, fib1+fib2

		if len(locatorHashes) > 1000 {
			break
		}
	}

	return locatorHashes
}
