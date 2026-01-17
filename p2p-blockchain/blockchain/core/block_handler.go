package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"

	"bjoernblessin.de/go-utils/util/assert"
	"bjoernblessin.de/go-utils/util/logger"
)

func (b *Blockchain) Block(receivedBlock block.Block, peerID common.PeerId) {
	logger.Infof("[block_handler] Block Message received from %v with %d transactions", peerID, len(receivedBlock.Transactions))

	// 1. Basic validation
	if ok, err := b.blockValidator.SanityCheck(receivedBlock); !ok {
		logger.Warnf("[block_handler] "+invalidBlockMessageFormat, peerID, err)
		return
	}

	if ok, err := b.blockValidator.ValidateHeaderOnly(receivedBlock.Header); !ok {
		logger.Warnf("[block_handler] "+invalidBlockMessageFormat, peerID, err)
		return
	}

	b.NotifyStopMining()
	// 2. Add block to store
	addedBlocks := b.blockStore.AddBlock(receivedBlock)

	// 3. Handle orphans
	if isOrphan, err := b.blockStore.IsOrphanBlock(receivedBlock); isOrphan {
		logger.Debugf("[block_handler] Block is Orphan: %v", err)
		assert.Assert(peerID != "", "Mined blocks should never be orphans")
		b.requestMissingBlockHeaders(receivedBlock, peerID)
		return
	}

	// 4. Full validation BEFORE applying to UTXO set
	if ok, _ := b.blockValidator.FullValidation(receivedBlock); !ok {
		// An invalid block will be "removed" in the block store in form of not beeing available for retreval
		return
	}

	// 5. Check if chain reorganization is needed
	tip := b.blockStore.GetMainChainTip()
	tipHash := tip.Hash()
	reorganized, err := b.chainReorganization.CheckAndReorganize(tipHash)
	if err != nil {
		logger.Errorf("[block_handler] Chain reorganization failed: %v", err)
		return
	}

	if reorganized {
		logger.Debugf("[block_handler] Chain reorganization performed")
	}

	// 6. Broadcast new blocks
	b.blockchainMsgSender.BroadcastAddedBlocks(addedBlocks, peerID)

	b.NotifyStartMining()
}

func (b *Blockchain) requestMissingBlockHeaders(receivedBlock block.Block, peerId common.PeerId) {
	parentHash := receivedBlock.Header.PreviousBlockHash

	currentHeight := b.blockStore.GetCurrentHeight()
	locatorHashes := b.buildBlockLocator(currentHeight)

	// Prepend the orphan parent hash at the beginning (most recent hash)
	locatorHashes = append([]common.Hash{parentHash}, locatorHashes...)

	locator := block.BlockLocator{
		BlockLocatorHashes: locatorHashes,
		StopHash:           common.Hash{}, // Empty stop hash means don't stop until we find common ancestor
	}

	b.blockchainMsgSender.RequestMissingBlockHeaders(locator, peerId)
}

// buildBlockLocator creates a block locator using Fibonacci series to sample the chain.
// Returns hashes starting from newer blocks (closer to tip) to older blocks (closer to genesis).
func (b *Blockchain) buildBlockLocator(tipHeight uint64) []common.Hash {
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
