package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"

	"bjoernblessin.de/go-utils/util/assert"
	"bjoernblessin.de/go-utils/util/logger"
)

const invalidBlockMessageFormat = "Block Message received from %v is invalid: %v"

// Block handles incoming block messages from the network or self-mined blocks.
// This is the main entry point for new blocks into the blockchain.
//
// Flow:
// 1. Basic validation (sanity check, header validation)
// 2. Add block to store
// 3. Handle orphan blocks
// 4. Full validation (PoW, merkle root)
// 5. UTXO validation using MultiChainUTXOService
// 6. Check for chain reorganization
// 7. Broadcast to peers
func (b *Blockchain) Block(receivedBlock block.Block, peerID common.PeerId) {
	logger.Infof("[block_handler] Block Message %v received from %v with %d transactions", &receivedBlock.Header,
		peerID, len(receivedBlock.Transactions))

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

	// 2. Handle orphans first - if parent doesn't exist, we can't validate UTXOs
	if isOrphan, err := b.blockStore.IsOrphanBlock(receivedBlock); isOrphan {
		// Add as orphan to store
		b.blockStore.AddBlock(receivedBlock)
		logger.Debugf("[block_handler] Block is Orphan %v  with error: %v", &receivedBlock.Header, err)
		assert.Assert(peerID != "", "Mined blocks should never be orphans")
		b.requestMissingBlockHeaders(receivedBlock, peerID)
		return
	}

	// 3. Full header validation (PoW, merkle root)
	if ok, err := b.blockValidator.FullValidation(receivedBlock); !ok {
		logger.Warnf(invalidBlockMessageFormat, peerID, err)
		return
	}
	// 6. Add block to store (after UTXO validation passed)
	addedBlocks := b.blockStore.AddBlock(receivedBlock)
	// 4. Determine if this is a main chain or side chain block
	blockHash := receivedBlock.Hash()
	parentHash := receivedBlock.Header.PreviousBlockHash
	mainChainTip := b.multiChainService.GetMainChainTip()

	// 5. UTXO validation and application
	if parentHash == mainChainTip {
		// Block extends main chain - validate and apply to main chain
		if err := b.handleMainChainBlock(receivedBlock); err != nil {
			logger.Warnf("Failed to handle main chain block %x: %v", blockHash[:4], err)
			return
		}
	} else {
		// Block creates or extends a side chain - validate using ephemeral view
		if err := b.handleSideChainBlock(receivedBlock); err != nil {
			logger.Warnf("Failed to handle side chain block %x: %v", blockHash[:4], err)
			return
		}
	}

	// 7. Check if chain reorganization is needed
	tip := b.blockStore.GetMainChainTip()
	tipHash := tip.Hash()
	reorganized, err := b.chainReorganization.CheckAndReorganize(tipHash)
	if err != nil {
		logger.Errorf("[block_handler] Chain reorganization failed: %v", err)
		return
	}

	if reorganized {
		logger.Debugf("[block_handler] Chain reorganization performed, new tip: %x", tipHash[:4])
	}

	// 8. Broadcast new blocks
	b.blockchainMsgSender.BroadcastAddedBlocks(addedBlocks, peerID)

	b.NotifyStartMining()
}

// handleMainChainBlock validates and applies a block that extends the main chain.
// The block's parent is the current main chain tip.
func (b *Blockchain) handleMainChainBlock(blk block.Block) error {
	blockHash := blk.Hash()
	// Current height before adding this block
	currentHeight := b.blockStore.GetCurrentHeight()
	// This block will be at height currentHeight + 1
	blockHeight := currentHeight + 1

	// Validate all transactions using an ephemeral UTXO view
	// This ensures inputs exist and aren't double-spent
	view, err := b.multiChainService.ValidateBlock(blk, blockHeight)
	if err != nil {
		logger.Warnf("Block %x failed UTXO validation: %v", blockHash[:4], err)
		return err
	}

	// Block is valid - apply transactions to main chain UTXO set
	for i, tx := range blk.Transactions {
		txID := tx.TransactionId()
		isCoinbase := i == 0 && tx.IsCoinbase()

		// Pass blockHeight (where this tx is included) and currentHeight (chain tip before this block)
		// The ApplyTransaction will determine if this is confirmed based on relative depth
		err := b.multiChainService.GetMainChain().ApplyTransaction(&tx, txID, blockHeight, blockHeight, isCoinbase)
		if err != nil {
			logger.Errorf("Failed to apply transaction %x to main chain: %v", txID[:4], err)
			return err
		}
	}

	// Update main chain tip
	b.multiChainService.SetMainChainTip(blockHash)

	// Remove confirmed transactions from mempool
	for _, tx := range blk.Transactions {
		if !tx.IsCoinbase() {
			b.mempool.RemoveTransaction(tx.TransactionId())
		}
	}

	logger.Debugf("Applied main chain block %x at height %d, view had %d added UTXOs",
		blockHash[:], blockHeight, len(view.GetAddedUTXOs()))

	return nil
}

// handleSideChainBlock validates a block that creates or extends a side chain.
// The block's parent is NOT the current main chain tip.
func (b *Blockchain) handleSideChainBlock(blk block.Block) error {
	blockHash := blk.Hash()
	parentHash := blk.Header.PreviousBlockHash

	// Calculate block height by walking back to find height
	blockHeight := b.calculateBlockHeight(parentHash) + 1

	// Validate and store delta for the side chain block
	delta, err := b.multiChainService.ValidateAndApplySideChainBlock(blk, blockHeight)
	if err != nil {
		logger.Warnf("Side chain block %x failed UTXO validation: %v", blockHash[:], err)
		return err
	}

	logger.Debugf("Validated side chain block %x at height %d, delta has %d added UTXOs, fork point: %x",
		blockHash[:], blockHeight, len(delta.AddedUTXOs), delta.ForkPoint[:])

	return nil
}

// calculateBlockHeight calculates the height of a block by walking back to genesis.
func (b *Blockchain) calculateBlockHeight(blockHash common.Hash) uint64 {
	height := uint64(0)
	currentHash := blockHash

	for {
		currentBlock, err := b.blockStore.GetBlockByHash(currentHash)
		if err != nil {
			break
		}

		if currentBlock.Header.PreviousBlockHash == (common.Hash{}) {
			// Reached genesis
			break
		}

		height++
		currentHash = currentBlock.Header.PreviousBlockHash
	}

	return height
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
