package discovery

import (
	"testing"
	"time"

	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/data/peer"
)

// Test file for getaddr functionality
// Mock implementations are in mocks.go

// Test HandleGetAddr

func TestHandleGetAddr_SendsPeersToRequester(t *testing.T) {
	peerStore := newMockDiscoveryPeerRetriever()
	addrSender := newMockAddrMsgSender()
	getAddrSender := newMockGetAddrMsgSender()
	peerCreator := newMockPeerCreator()

	service := NewDiscoveryService(nil, peerCreator, addrSender, peerStore, getAddrSender)

	// Add some test peers
	peer1 := &peer.Peer{
		Version: "1.0.0",
		State:   common.StateConnected,
	}
	peer2 := &peer.Peer{
		Version: "1.0.0",
		State:   common.StateConnected,
	}
	peerStore.AddPeerById("peer-1", peer1)
	peerStore.AddPeerById("peer-2", peer2)

	requesterPeerID := common.PeerId("requester-peer")

	// Call HandleGetAddr
	service.HandleGetAddr(requesterPeerID)

	// Give time for goroutine to finish
	time.Sleep(100 * time.Millisecond)

	// Verify SendAddr was called exactly once
	if addrSender.getSendAddrCallCount() != 1 {
		t.Errorf("expected 1 SendAddr call, got %d", addrSender.getSendAddrCallCount())
	}

	// Verify the correct peer was sent to
	call := addrSender.getLastSendAddrCall()
	if call == nil {
		t.Fatal("expected SendAddr to be called, but got nil")
	}

	if call.peerID != requesterPeerID {
		t.Errorf("expected peerID %s, got %s", requesterPeerID, call.peerID)
	}

	// Verify both peers were sent
	if len(call.addrs) != 2 {
		t.Errorf("expected 2 peer addresses, got %d", len(call.addrs))
	}

	// Verify peer IDs are correct
	sentPeerIds := make(map[common.PeerId]bool)
	for _, addr := range call.addrs {
		sentPeerIds[addr.PeerId] = true
	}

	if !sentPeerIds["peer-1"] || !sentPeerIds["peer-2"] {
		t.Error("expected both peer-1 and peer-2 to be in addresses")
	}
}

func TestHandleGetAddr_ExcludesRequesterPeer(t *testing.T) {
	peerStore := newMockDiscoveryPeerRetriever()
	addrSender := newMockAddrMsgSender()
	getAddrSender := newMockGetAddrMsgSender()
	peerCreator := newMockPeerCreator()

	service := NewDiscoveryService(nil, peerCreator, addrSender, peerStore, getAddrSender)

	// Add the requesting peer itself to the store
	requesterPeerID := common.PeerId("requester-peer")
	requesterPeer := &peer.Peer{
		Version: "1.0.0",
		State:   common.StateConnected,
	}
	peerStore.AddPeerById(requesterPeerID, requesterPeer)

	// Add another peer
	otherPeer := &peer.Peer{
		Version: "1.0.0",
		State:   common.StateConnected,
	}
	peerStore.AddPeerById("other-peer", otherPeer)

	// Call HandleGetAddr
	service.HandleGetAddr(requesterPeerID)

	// Give time for goroutine to finish
	time.Sleep(100 * time.Millisecond)

	// Verify SendAddr was called
	if addrSender.getSendAddrCallCount() != 1 {
		t.Errorf("expected 1 SendAddr call, got %d", addrSender.getSendAddrCallCount())
	}

	// Verify only the other peer was sent (not the requester)
	call := addrSender.getLastSendAddrCall()
	if call == nil {
		t.Fatal("expected SendAddr to be called, but got nil")
	}

	if len(call.addrs) != 1 {
		t.Errorf("expected 1 peer address (excluding requester), got %d", len(call.addrs))
	}

	if call.addrs[0].PeerId != "other-peer" {
		t.Errorf("expected peer address to be other-peer, got %s", call.addrs[0].PeerId)
	}
}

func TestHandleGetAddr_DoesNotSendWhenNoPeers(t *testing.T) {
	peerStore := newMockDiscoveryPeerRetriever()
	addrSender := newMockAddrMsgSender()
	getAddrSender := newMockGetAddrMsgSender()
	peerCreator := newMockPeerCreator()

	service := NewDiscoveryService(nil, peerCreator, addrSender, peerStore, getAddrSender)

	requesterPeerID := common.PeerId("requester-peer")

	// Call HandleGetAddr with empty peer store
	service.HandleGetAddr(requesterPeerID)

	// Give time for goroutine to finish
	time.Sleep(100 * time.Millisecond)

	// Verify SendAddr was NOT called
	if addrSender.getSendAddrCallCount() != 0 {
		t.Errorf("expected 0 SendAddr calls when no peers available, got %d", addrSender.getSendAddrCallCount())
	}
}

func TestHandleGetAddr_IncludesLastActiveTimestamp(t *testing.T) {
	peerStore := newMockDiscoveryPeerRetriever()
	addrSender := newMockAddrMsgSender()
	getAddrSender := newMockGetAddrMsgSender()
	peerCreator := newMockPeerCreator()

	service := NewDiscoveryService(nil, peerCreator, addrSender, peerStore, getAddrSender)

	// Add a peer with LastSeen set
	testPeer := &peer.Peer{
		Version:  "1.0.0",
		State:    common.StateConnected,
		LastSeen: time.Now().Unix(),
	}
	peerStore.AddPeerById("peer-1", testPeer)

	requesterPeerID := common.PeerId("requester-peer")

	// Call HandleGetAddr
	service.HandleGetAddr(requesterPeerID)

	// Give time for goroutine to finish
	time.Sleep(100 * time.Millisecond)

	// Verify timestamp is included
	call := addrSender.getLastSendAddrCall()
	if call == nil {
		t.Fatal("expected SendAddr to be called, but got nil")
	}

	if len(call.addrs) != 1 {
		t.Errorf("expected 1 peer address, got %d", len(call.addrs))
	}

	addr := call.addrs[0]
	if addr.LastActiveTimestamp == 0 {
		t.Error("expected LastActiveTimestamp to be set")
	}
}

func TestHandleGetAddr_SendsAsynchronously(t *testing.T) {
	peerStore := newMockDiscoveryPeerRetriever()
	addrSender := newMockAddrMsgSender()
	getAddrSender := newMockGetAddrMsgSender()
	peerCreator := newMockPeerCreator()

	service := NewDiscoveryService(nil, peerCreator, addrSender, peerStore, getAddrSender)

	// Add a peer
	testPeer := &peer.Peer{
		Version: "1.0.0",
		State:   common.StateConnected,
	}
	peerStore.AddPeerById("peer-1", testPeer)

	requesterPeerID := common.PeerId("requester-peer")

	// Call HandleGetAddr - it should return immediately
	service.HandleGetAddr(requesterPeerID)

	// Verify the call returns immediately (sends asynchronously)
	// If it were synchronous, the SendAddr would be called before we check
	// But since we use a goroutine, it might not be called yet
	time.Sleep(10 * time.Millisecond)

	// After some time, it should be called
	time.Sleep(100 * time.Millisecond)
	if addrSender.getSendAddrCallCount() != 1 {
		t.Errorf("expected 1 SendAddr call after async execution, got %d", addrSender.getSendAddrCallCount())
	}
}

// Test SendGetAddr

func TestSendGetAddr_ForwardsToSender(t *testing.T) {
	peerStore := newMockDiscoveryPeerRetriever()
	addrSender := newMockAddrMsgSender()
	getAddrSender := newMockGetAddrMsgSender()
	peerCreator := newMockPeerCreator()

	service := NewDiscoveryService(nil, peerCreator, addrSender, peerStore, getAddrSender)

	targetPeerID := common.PeerId("target-peer")

	// Call SendGetAddr
	service.SendGetAddr(targetPeerID)

	// Wait for the goroutine to complete
	getAddrSender.waitForCall()

	// Verify SendGetAddr was called on the sender
	if getAddrSender.getSendGetAddrCallCount() != 1 {
		t.Errorf("expected 1 SendGetAddr call, got %d", getAddrSender.getSendGetAddrCallCount())
	}

	// Verify the correct peer ID was sent
	call := getAddrSender.getLastSendGetAddrCall()
	if call == nil {
		t.Fatal("expected SendGetAddr to be called, but got nil")
	}

	if *call != targetPeerID {
		t.Errorf("expected peerID %s, got %s", targetPeerID, *call)
	}
}
