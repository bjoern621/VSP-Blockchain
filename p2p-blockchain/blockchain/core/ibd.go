package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"

	"bjoernblessin.de/go-utils/util/logger"
)

// OnPeerConnected implements the ConnectionObserver interface.
// This is called when a peer's handshake completes.
// IBD is only triggered for outbound connections (isOutbound=true) to avoid duplicate syncs.
// Following Bitcoin's Headers-First approach:
// 1. Send GetHeaders with a block locator to discover new headers
// 2. The peer responds with Headers message (handled by Headers() method)
// 3. For unknown headers, we request full blocks via GetData (done in headers_handler.go)
// 4. Full blocks are received and processed (done in block_handler.go)
func (b *Blockchain) OnPeerConnected(peerID common.PeerId, isOutbound bool) {
	if !isOutbound {
		logger.Debugf("[ibd] Peer %s connected (inbound) - not initiating IBD", peerID)
		return
	}

	logger.Infof("[ibd] Peer %s connected (outbound), initiating Initial Block Download", peerID)

	// Build block locator from current chain state
	currentHeight := b.blockStore.GetMainChainHeight()
	locatorHashes := b.buildBlockLocator(currentHeight)

	// Create block locator for GetHeaders request
	// StopHash is empty to get all headers the peer has
	locator := block.BlockLocator{
		BlockLocatorHashes: locatorHashes,
		StopHash:           common.Hash{}, // Empty means "don't stop, give me everything you have"
	}

	logger.Infof("[ibd] Requesting headers from peer %s, current height: %d, locator size: %d",
		peerID, currentHeight, len(locatorHashes))

	// Request missing block headers from the newly connected peer
	b.blockchainMsgSender.RequestMissingBlockHeaders(locator, peerID)
}
