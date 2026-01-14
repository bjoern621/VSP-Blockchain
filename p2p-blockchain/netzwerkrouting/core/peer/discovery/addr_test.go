package discovery

import (
	"testing"
	"time"

	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/data/peer"
)

// Test file for addr functionality
// Mock implementations are in mocks.go

// Test HandleAddr

func TestHandleAddr_UpdatesPeerLastSeenTimestamp(t *testing.T) {
	peerStore := newMockDiscoveryPeerRetriever()
	addrSender := newMockAddrMsgSender()
	getAddrSender := newMockGetAddrMsgSender()
	peerCreator := newMockPeerCreator()

	service := NewDiscoveryService(nil, peerCreator, addrSender, peerStore, getAddrSender)

	// Create and add a peer with an old timestamp
	oldTimestamp := time.Now().Add(-24 * time.Hour).Unix()
	testPeer := &peer.Peer{
		Version:  "1.0.0",
		State:    common.StateConnected,
		LastSeen: oldTimestamp,
	}
	peerStore.AddPeerById("peer-1", testPeer)

	senderPeerID := common.PeerId("sender-peer")
	newTimestamp := time.Now().Unix()

	// Call HandleAddr with a newer timestamp
	addresses := []PeerAddress{
		{
			PeerId:              "peer-1",
			LastActiveTimestamp: newTimestamp,
		},
	}

	service.HandleAddr(senderPeerID, addresses)

	// Verify the peer's LastSeen was updated
	retrievedPeer, exists := peerStore.GetPeer("peer-1")
	if !exists {
		t.Fatal("expected peer to exist")
	}

	if retrievedPeer.LastSeen != newTimestamp {
		t.Errorf("expected LastSeen to be updated to %d, got %d", newTimestamp, retrievedPeer.LastSeen)
	}
}

func TestHandleAddr_DoesNotUpdateOlderTimestamp(t *testing.T) {
	peerStore := newMockDiscoveryPeerRetriever()
	addrSender := newMockAddrMsgSender()
	getAddrSender := newMockGetAddrMsgSender()
	peerCreator := newMockPeerCreator()

	service := NewDiscoveryService(nil, peerCreator, addrSender, peerStore, getAddrSender)

	// Create and add a peer with a recent timestamp
	recentTimestamp := time.Now().Unix()
	testPeer := &peer.Peer{
		Version:  "1.0.0",
		State:    common.StateConnected,
		LastSeen: recentTimestamp,
	}
	peerStore.AddPeerById("peer-1", testPeer)

	senderPeerID := common.PeerId("sender-peer")
	oldTimestamp := time.Now().Add(-24 * time.Hour).Unix()

	// Call HandleAddr with an older timestamp
	addresses := []PeerAddress{
		{
			PeerId:              "peer-1",
			LastActiveTimestamp: oldTimestamp,
		},
	}

	service.HandleAddr(senderPeerID, addresses)

	// Verify the peer's LastSeen was NOT updated (kept the recent timestamp)
	retrievedPeer, exists := peerStore.GetPeer("peer-1")
	if !exists {
		t.Fatal("expected peer to exist")
	}

	if retrievedPeer.LastSeen != recentTimestamp {
		t.Errorf("expected LastSeen to remain %d, got %d", recentTimestamp, retrievedPeer.LastSeen)
	}
}

func TestHandleAddr_HandlesMultipleAddresses(t *testing.T) {
	peerStore := newMockDiscoveryPeerRetriever()
	addrSender := newMockAddrMsgSender()
	getAddrSender := newMockGetAddrMsgSender()
	peerCreator := newMockPeerCreator()

	service := NewDiscoveryService(nil, peerCreator, addrSender, peerStore, getAddrSender)

	// Create and add multiple peers
	now := time.Now().Unix()
	peer1 := &peer.Peer{
		Version:  "1.0.0",
		State:    common.StateConnected,
		LastSeen: now - 3600,
	}
	peer2 := &peer.Peer{
		Version:  "1.0.0",
		State:    common.StateConnected,
		LastSeen: now - 7200,
	}
	peer3 := &peer.Peer{
		Version:  "1.0.0",
		State:    common.StateConnected,
		LastSeen: now - 1800,
	}
	peerStore.AddPeerById("peer-1", peer1)
	peerStore.AddPeerById("peer-2", peer2)
	peerStore.AddPeerById("peer-3", peer3)

	senderPeerID := common.PeerId("sender-peer")

	// Call HandleAddr with multiple addresses
	addresses := []PeerAddress{
		{
			PeerId:              "peer-1",
			LastActiveTimestamp: now,
		},
		{
			PeerId:              "peer-2",
			LastActiveTimestamp: now - 100,
		},
		{
			PeerId:              "peer-3",
			LastActiveTimestamp: now,
		},
	}

	service.HandleAddr(senderPeerID, addresses)

	// Verify all peers were updated correctly
	for _, addr := range addresses {
		retrievedPeer, exists := peerStore.GetPeer(addr.PeerId)
		if !exists {
			t.Errorf("expected peer %s to exist", addr.PeerId)
			continue
		}

		if retrievedPeer.LastSeen != addr.LastActiveTimestamp {
			t.Errorf("expected LastSeen for peer %s to be %d, got %d",
				addr.PeerId, addr.LastActiveTimestamp, retrievedPeer.LastSeen)
		}
	}
}

func TestHandleAddr_HandlesEmptyAddressList(t *testing.T) {
	peerStore := newMockDiscoveryPeerRetriever()
	addrSender := newMockAddrMsgSender()
	getAddrSender := newMockGetAddrMsgSender()
	peerCreator := newMockPeerCreator()

	service := NewDiscoveryService(nil, peerCreator, addrSender, peerStore, getAddrSender)

	// Add a peer
	now := time.Now().Unix()
	testPeer := &peer.Peer{
		Version:  "1.0.0",
		State:    common.StateConnected,
		LastSeen: now,
	}
	peerStore.AddPeerById("peer-1", testPeer)

	senderPeerID := common.PeerId("sender-peer")

	// Call HandleAddr with empty address list
	addresses := []PeerAddress{}

	// This should not panic
	service.HandleAddr(senderPeerID, addresses)

	// Peer should remain unchanged
	retrievedPeer, exists := peerStore.GetPeer("peer-1")
	if !exists {
		t.Fatal("expected peer to exist")
	}

	if retrievedPeer.LastSeen != now {
		t.Error("expected peer LastSeen to remain unchanged")
	}
}

func TestHandleAddr_ThreadSafe(t *testing.T) {
	peerStore := newMockDiscoveryPeerRetriever()
	addrSender := newMockAddrMsgSender()
	getAddrSender := newMockGetAddrMsgSender()
	peerCreator := newMockPeerCreator()

	service := NewDiscoveryService(nil, peerCreator, addrSender, peerStore, getAddrSender)

	// Create a peer
	now := time.Now().Unix()
	testPeer := &peer.Peer{
		Version:  "1.0.0",
		State:    common.StateConnected,
		LastSeen: now,
	}
	peerStore.AddPeerById("peer-1", testPeer)

	senderPeerID := common.PeerId("sender-peer")
	newTimestamp := now + 3600

	// Call HandleAddr from multiple goroutines concurrently
	addresses := []PeerAddress{
		{
			PeerId:              "peer-1",
			LastActiveTimestamp: newTimestamp,
		},
	}

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			service.HandleAddr(senderPeerID, addresses)
			done <- true
		}()
	}

	// Wait for all goroutines to finish
	for i := 0; i < 10; i++ {
		<-done
	}

	// Peer should still be in a consistent state
	retrievedPeer, exists := peerStore.GetPeer("peer-1")
	if !exists {
		t.Fatal("expected peer to exist after concurrent updates")
	}

	// LastSeen should be the latest timestamp
	if retrievedPeer.LastSeen != newTimestamp {
		t.Errorf("expected LastSeen to be %d, got %d", newTimestamp, retrievedPeer.LastSeen)
	}
}

func TestHandleAddr_PanicsIfPeerNotRegistered(t *testing.T) {
	peerStore := newMockDiscoveryPeerRetriever()
	addrSender := newMockAddrMsgSender()
	getAddrSender := newMockGetAddrMsgSender()
	peerCreator := newMockPeerCreator()

	service := NewDiscoveryService(nil, peerCreator, addrSender, peerStore, getAddrSender)

	// Don't add any peer to the store

	senderPeerID := common.PeerId("sender-peer")
	now := time.Now().Unix()

	// Call HandleAddr with a peer that doesn't exist
	addresses := []PeerAddress{
		{
			PeerId:              "non-existent-peer",
			LastActiveTimestamp: now,
		},
	}

	// This should panic because the peer is not registered
	// (as per the assertion in the code)
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected HandleAddr to panic when peer is not registered")
		}
	}()

	service.HandleAddr(senderPeerID, addresses)
}

func TestHandleAddr_UpdatesCorrectPeer(t *testing.T) {
	peerStore := newMockDiscoveryPeerRetriever()
	addrSender := newMockAddrMsgSender()
	getAddrSender := newMockGetAddrMsgSender()
	peerCreator := newMockPeerCreator()

	service := NewDiscoveryService(nil, peerCreator, addrSender, peerStore, getAddrSender)

	// Create and add multiple peers with different timestamps
	now := time.Now().Unix()

	peer1 := &peer.Peer{
		Version:  "1.0.0",
		State:    common.StateConnected,
		LastSeen: now - 3600,
	}
	peer2 := &peer.Peer{
		Version:  "1.0.0",
		State:    common.StateConnected,
		LastSeen: now - 7200,
	}
	peer3 := &peer.Peer{
		Version:  "1.0.0",
		State:    common.StateConnected,
		LastSeen: now - 1800,
	}
	peerStore.AddPeerById("peer-1", peer1)
	peerStore.AddPeerById("peer-2", peer2)
	peerStore.AddPeerById("peer-3", peer3)

	senderPeerID := common.PeerId("sender-peer")

	// Call HandleAddr with only peer-2's address updated
	newTimestamp := now
	addresses := []PeerAddress{
		{
			PeerId:              "peer-2",
			LastActiveTimestamp: newTimestamp,
		},
	}

	service.HandleAddr(senderPeerID, addresses)

	// Verify only peer-2 was updated
	for peerID, expectedLastSeen := range map[common.PeerId]int64{
		"peer-1": now - 3600,
		"peer-2": newTimestamp,
		"peer-3": now - 1800,
	} {
		retrievedPeer, exists := peerStore.GetPeer(peerID)
		if !exists {
			t.Errorf("expected peer %s to exist", peerID)
			continue
		}

		if retrievedPeer.LastSeen != expectedLastSeen {
			t.Errorf("expected LastSeen for peer %s to be %d, got %d",
				peerID, expectedLastSeen, retrievedPeer.LastSeen)
		}
	}
}

func TestHandleAddr_SameTimestamp(t *testing.T) {
	peerStore := newMockDiscoveryPeerRetriever()
	addrSender := newMockAddrMsgSender()
	getAddrSender := newMockGetAddrMsgSender()
	peerCreator := newMockPeerCreator()

	service := NewDiscoveryService(nil, peerCreator, addrSender, peerStore, getAddrSender)

	// Create and add a peer
	timestamp := time.Now().Unix()
	testPeer := &peer.Peer{
		Version:  "1.0.0",
		State:    common.StateConnected,
		LastSeen: timestamp,
	}
	peerStore.AddPeerById("peer-1", testPeer)

	senderPeerID := common.PeerId("sender-peer")

	// Call HandleAddr with the same timestamp
	addresses := []PeerAddress{
		{
			PeerId:              "peer-1",
			LastActiveTimestamp: timestamp,
		},
	}

	service.HandleAddr(senderPeerID, addresses)

	// Verify the peer's LastSeen remains the same
	retrievedPeer, exists := peerStore.GetPeer("peer-1")
	if !exists {
		t.Fatal("expected peer to exist")
	}

	if retrievedPeer.LastSeen != timestamp {
		t.Errorf("expected LastSeen to remain %d, got %d", timestamp, retrievedPeer.LastSeen)
	}
}
