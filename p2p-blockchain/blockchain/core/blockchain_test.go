package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/validation"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/inv"
	"testing"

	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
)

type mockBlockchainSender struct {
	called     bool
	lastMsg    []*inv.InvVector
	lastPeerID common.PeerId
	callsCount int
}

func (m *mockBlockchainSender) SendGetData(msg []*inv.InvVector, peerId common.PeerId) {
	m.called = true
	m.callsCount++
	m.lastMsg = msg
	m.lastPeerID = peerId
}

func (m *mockBlockchainSender) BroadcastInvExclusionary(msg []*inv.InvVector, peerId common.PeerId) {}

func TestBlockchain_Inv_InvokesRequestDataByCallingSendGetData(t *testing.T) {
	// Arrange: create blockchain with mocked sender
	sender := &mockBlockchainSender{}
	bc := NewBlockchain(sender, validation.ValidationService{})

	peerID := common.PeerId("peer-1")

	var h common.Hash
	h[0] = 0xAB // arbitrary non-zero hash to make assertions clearer

	invVector := &inv.InvVector{
		InvType: inv.InvTypeMsgTx,
		Hash:    h,
	}
	inventory := []*inv.InvVector{
		invVector,
	}
	// Act: receive Inv message
	bc.Inv(inventory, peerID)

	// Assert: requestData() is exercised by observing its effect: SendGetData is invoked
	if !sender.called {
		t.Fatalf("expected SendGetData to be called (requestData invoked), but it was not called")
	}
	if sender.callsCount != 1 {
		t.Fatalf("expected SendGetData to be called once, but was called %d times", sender.callsCount)
	}
	if sender.lastPeerID != peerID {
		t.Fatalf("expected peerID %q, got %q", peerID, sender.lastPeerID)
	}
	if len(sender.lastMsg) != 1 {
		t.Fatalf("expected GetData inventory length 1, got %d", len(sender.lastMsg))
	}
	if sender.lastMsg[0].InvType != inv.InvTypeMsgTx {
		t.Fatalf("expected inventory[0].Type %v, got %v", inv.InvTypeMsgTx, sender.lastMsg[0].InvType)
	}
	if sender.lastMsg[0].Hash != h {
		t.Fatalf("expected inventory[0].Hash %v, got %v", h, sender.lastMsg[0].Hash)
	}
}
