package core

import (
	"testing"

	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/validation"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
)

func TestOnPeerConnected_OutboundSendsGetHeaders(t *testing.T) {
	// Arrange
	mockMsgSender := &mockBlockchainSender{}

	genesisBlock := createTestBlock(common.Hash{}, 0)
	mockBlockStore := &mockBlockStore{
		currentHeight: 0,
		mainChainTip:  genesisBlock,
		blocksByHeight: map[uint64][]block.Block{
			0: {genesisBlock},
		},
		isPartOfMainChain: func(_ block.Block) bool { return true },
	}

	blockchain := NewBlockchain(
		mockMsgSender,
		nil,
		validation.NewTransactionValidator(nil),
		nil,
		mockBlockStore,
		nil,
	)

	testPeerID := common.PeerId("test-peer-123")

	// Act: trigger IBD by simulating outbound peer connection
	blockchain.OnPeerConnected(testPeerID, true) // isOutbound=true

	// Assert: GetHeaders was called via RequestMissingBlockHeaders
	if !mockMsgSender.requestMissingHeadersCalled {
		t.Fatal("expected RequestMissingBlockHeaders to be called for outbound connection")
	}

	if mockMsgSender.requestMissingHeadersPeerId != testPeerID {
		t.Errorf("expected peerID %s, got %s", testPeerID, mockMsgSender.requestMissingHeadersPeerId)
	}

	// Verify block locator contains at least the genesis block hash
	locator := mockMsgSender.requestMissingHeadersLocator
	if len(locator.BlockLocatorHashes) == 0 {
		t.Error("expected block locator to contain at least one hash")
	}

	// Verify stop hash is empty (request all headers)
	if locator.StopHash != (common.Hash{}) {
		t.Error("expected empty stop hash for full header sync")
	}
}

func TestOnPeerConnected_InboundDoesNotTriggerIBD(t *testing.T) {
	// Arrange
	mockMsgSender := &mockBlockchainSender{}

	genesisBlock := createTestBlock(common.Hash{}, 0)
	mockBlockStore := &mockBlockStore{
		currentHeight: 0,
		mainChainTip:  genesisBlock,
		blocksByHeight: map[uint64][]block.Block{
			0: {genesisBlock},
		},
		isPartOfMainChain: func(_ block.Block) bool { return true },
	}

	blockchain := NewBlockchain(
		mockMsgSender,
		nil,
		validation.NewTransactionValidator(nil),
		nil,
		mockBlockStore,
		nil,
	)

	testPeerID := common.PeerId("test-peer-123")

	// Act: simulate inbound peer connection
	blockchain.OnPeerConnected(testPeerID, false) // isOutbound=false

	// Assert: GetHeaders should NOT be called for inbound connections
	if mockMsgSender.requestMissingHeadersCalled {
		t.Fatal("expected RequestMissingBlockHeaders NOT to be called for inbound connection")
	}
}

func TestOnPeerConnected_BlockLocatorUsesFibonacci(t *testing.T) {
	// Arrange
	mockMsgSender := &mockBlockchainSender{}

	// Create a mock block store with multiple blocks
	blocksByHeight := make(map[uint64][]block.Block)
	for i := uint64(0); i <= 10; i++ {
		blocksByHeight[i] = []block.Block{createTestBlock(common.Hash{byte(i)}, uint32(i))}
	}

	mockBlockStore := &mockBlockStore{
		currentHeight:     10,
		mainChainTip:      blocksByHeight[10][0],
		blocksByHeight:    blocksByHeight,
		isPartOfMainChain: func(_ block.Block) bool { return true },
	}

	blockchain := NewBlockchain(
		mockMsgSender,
		nil,
		validation.NewTransactionValidator(nil),
		nil,
		mockBlockStore,
		nil,
	)

	// Act - outbound connection triggers IBD
	blockchain.OnPeerConnected("test-peer", true)

	// Assert
	if !mockMsgSender.requestMissingHeadersCalled {
		t.Fatal("expected RequestMissingBlockHeaders to be called")
	}

	locator := mockMsgSender.requestMissingHeadersLocator

	// With height 10 and Fibonacci sampling (offsets: 0, 1, 2, 3, 5, 8, 13...)
	// We should get hashes at heights: 10, 9, 8, 7, 5, 2 (stopping when offset > tipHeight)
	if len(locator.BlockLocatorHashes) < 3 {
		t.Errorf("expected at least 3 block locator hashes for height 10, got %d", len(locator.BlockLocatorHashes))
	}
}
