package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/validation"
	"testing"

	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api/blockchain/dto"
)

type mockBlockchainSender struct {
	called     bool
	lastMsg    dto.GetDataMsgDTO
	lastPeerID common.PeerId
	callsCount int
}

func (m *mockBlockchainSender) SendGetData(msg dto.GetDataMsgDTO, peerId common.PeerId) {
	m.called = true
	m.callsCount++
	m.lastMsg = msg
	m.lastPeerID = peerId
}

func (m *mockBlockchainSender) BroadcastInv(msg dto.InvMsgDTO, peerId common.PeerId) {}

func TestBlockchain_Inv_InvokesRequestDataByCallingSendGetData(t *testing.T) {
	// Arrange: create blockchain with mocked sender
	sender := &mockBlockchainSender{}
	bc := NewBlockchain(sender, validation.ValidationService{})

	peerID := common.PeerId("peer-1")

	var h dto.Hash
	h[0] = 0xAB // arbitrary non-zero hash to make assertions clearer

	invMsg := dto.InvMsgDTO{
		Inventory: []dto.InvVectorDTO{
			{
				Type: dto.InvTypeDTO_MSG_TX,
				Hash: h,
			},
		},
	}

	// Act: receive Inv message
	bc.Inv(invMsg, peerID)

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
	if len(sender.lastMsg.Inventory) != 1 {
		t.Fatalf("expected GetData inventory length 1, got %d", len(sender.lastMsg.Inventory))
	}
	if sender.lastMsg.Inventory[0].Type != dto.InvTypeDTO_MSG_TX {
		t.Fatalf("expected inventory[0].Type %v, got %v", dto.InvTypeDTO_MSG_TX, sender.lastMsg.Inventory[0].Type)
	}
	if sender.lastMsg.Inventory[0].Hash != h {
		t.Fatalf("expected inventory[0].Hash %v, got %v", h, sender.lastMsg.Inventory[0].Hash)
	}
}
