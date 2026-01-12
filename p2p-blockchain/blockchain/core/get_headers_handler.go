package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"

	"bjoernblessin.de/go-utils/util/logger"
)

func (b *Blockchain) GetHeaders(locator block.BlockLocator, peerID common.PeerId) {
	logger.Infof("GetHeaders Message received: %v from %v", locator, peerID)

	// Find the common ancestor by checking the block locator hashes
	// The locator hashes are ordered from newest to oldest
	var commonAncestorHash common.Hash
	var commonAncestorHeight uint64

	for _, hash := range locator.BlockLocatorHashes {
		// Check if this hash exists in our block store
		if block, err := b.blockStore.GetBlockByHash(hash); err == nil {
			// Found a block that exists in our chain - this is a potential common ancestor
			// Verify it's on our main chain
			if b.blockStore.IsPartOfMainChain(block) {
				commonAncestorHash = hash
				commonAncestorHeight = b.findBlockHeight(hash)
				break
			}
		}
	}

	// If no common ancestor found in locator, use genesis
	if commonAncestorHash == (common.Hash{}) {
		tip := b.blockStore.GetMainChainTip()
		genesisBlock, err := b.blockStore.GetBlockByHash(tip.Hash())
		// Traverse to genesis
		for err == nil && genesisBlock.Header.PreviousBlockHash != (common.Hash{}) {
			genesisBlock, err = b.blockStore.GetBlockByHash(genesisBlock.Header.PreviousBlockHash)
		}
		if err == nil {
			commonAncestorHeight = 0
		} else {
			logger.Warnf("Failed to find genesis block for GetHeaders response")
			return
		}
	}

	// Collect headers starting from the block after the common ancestor
	// Maximum 100 headers as per protocol specification
	headers := b.collectBlockHeaders(locator, commonAncestorHeight)

	// Send headers back to the requesting peer
	b.sendHeadersBackToPeer(headers, peerID, commonAncestorHeight)
}

func (b *Blockchain) collectBlockHeaders(locator block.BlockLocator, commonAncestorHeight uint64) []*block.BlockHeader {
	const maxHeaders = 100
	headers := make([]*block.BlockHeader, 0, maxHeaders)

	currentHeight := commonAncestorHeight + 1
	currentTipHeight := b.blockStore.GetCurrentHeight()

	for len(headers) < maxHeaders && currentHeight <= currentTipHeight {
		blocksAtHeight := b.blockStore.GetBlocksByHeight(currentHeight)
		if len(blocksAtHeight) == 0 {
			break
		}

		// Get the main chain block at this height
		var mainChainBlock *block.Block
		for _, blk := range blocksAtHeight {
			if b.blockStore.IsPartOfMainChain(blk) {
				mainChainBlock = &blk
				break
			}
		}

		if mainChainBlock != nil {
			headers = append(headers, &mainChainBlock.Header)

			// Check if we've reached the stop hash
			if locator.StopHash == mainChainBlock.Hash() {
				break
			}
		}

		currentHeight++
	}
	return headers
}

func (b *Blockchain) sendHeadersBackToPeer(headers []*block.BlockHeader, peerID common.PeerId, commonAncestorHeight uint64) {
	if len(headers) > 0 {
		logger.Infof("Sending %d headers to peer %s, starting from height %d", len(headers), peerID, commonAncestorHeight+1)
		b.blockchainMsgSender.SendHeaders(headers, peerID)
	} else {
		logger.Infof("No headers to send to peer %s", peerID)
	}
}

// findBlockHeight finds the height of a block with the given hash
func (b *Blockchain) findBlockHeight(hash common.Hash) uint64 {
	// Start from the tip and work backwards
	tip := b.blockStore.GetMainChainTip()
	tipHash := tip.Hash()
	currentTipHeight := b.blockStore.GetCurrentHeight()

	// If we're looking for the tip, return current height
	if hash == tipHash {
		return currentTipHeight
	}

	// Traverse backwards from tip
	currentBlock := tip
	height := currentTipHeight

	for height > 0 {
		if currentBlock.Hash() == hash {
			return height
		}
		// Move to parent
		parentHash := currentBlock.Header.PreviousBlockHash
		parentBlock, err := b.blockStore.GetBlockByHash(parentHash)
		if err != nil {
			break
		}
		currentBlock = parentBlock
		height--
	}

	// Check genesis (height 0)
	if currentBlock.Hash() == hash {
		return 0
	}

	return 0
}
