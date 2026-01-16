package discovery

import (
	"testing"
	"time"

	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/data/peer"

	mapset "github.com/deckarep/golang-set/v2"
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
		Version:     "1.0.0",
		State:       common.StateConnected,
		LastSeen:    oldTimestamp,
		AddrsSentTo: mapset.NewSet[common.PeerId](),
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
		Version:     "1.0.0",
		State:       common.StateConnected,
		LastSeen:    recentTimestamp,
		AddrsSentTo: mapset.NewSet[common.PeerId](),
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
		Version:     "1.0.0",
		State:       common.StateConnected,
		LastSeen:    now - 3600,
		AddrsSentTo: mapset.NewSet[common.PeerId](),
	}
	peer2 := &peer.Peer{
		Version:     "1.0.0",
		State:       common.StateConnected,
		LastSeen:    now - 7200,
		AddrsSentTo: mapset.NewSet[common.PeerId](),
	}
	peer3 := &peer.Peer{
		Version:     "1.0.0",
		State:       common.StateConnected,
		LastSeen:    now - 1800,
		AddrsSentTo: mapset.NewSet[common.PeerId](),
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
		Version:     "1.0.0",
		State:       common.StateConnected,
		LastSeen:    now,
		AddrsSentTo: mapset.NewSet[common.PeerId](),
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
		Version:     "1.0.0",
		State:       common.StateConnected,
		LastSeen:    now,
		AddrsSentTo: mapset.NewSet[common.PeerId](),
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
		Version:     "1.0.0",
		State:       common.StateConnected,
		LastSeen:    now - 3600,
		AddrsSentTo: mapset.NewSet[common.PeerId](),
	}
	peer2 := &peer.Peer{
		Version:     "1.0.0",
		State:       common.StateConnected,
		LastSeen:    now - 7200,
		AddrsSentTo: mapset.NewSet[common.PeerId](),
	}
	peer3 := &peer.Peer{
		Version:     "1.0.0",
		State:       common.StateConnected,
		LastSeen:    now - 1800,
		AddrsSentTo: mapset.NewSet[common.PeerId](),
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
		Version:     "1.0.0",
		State:       common.StateConnected,
		LastSeen:    timestamp,
		AddrsSentTo: mapset.NewSet[common.PeerId](),
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

// Tests for forwarding addresses

// TestForwardAddrs_DoesNotForwardToSender verifies that addresses are not forwarded
// back to the peer from which they were received.
func TestForwardAddrs_DoesNotForwardToSender(t *testing.T) {
	peerStore := newMockDiscoveryPeerRetriever()
	addrSender := newMockAddrMsgSender()
	getAddrSender := newMockGetAddrMsgSender()
	peerCreator := newMockPeerCreator()

	service := NewDiscoveryService(nil, peerCreator, addrSender, peerStore, getAddrSender)

	// Create and add connected peers including the sender
	connectedPeers := []common.PeerId{"sender-peer", "peer-1", "peer-2", "peer-3"}
	now := time.Now().Unix()

	for _, peerID := range connectedPeers {
		testPeer := &peer.Peer{
			Version:     "1.0.0",
			State:       common.StateConnected,
			LastSeen:    now,
			AddrsSentTo: mapset.NewSet[common.PeerId](),
		}
		peerStore.AddPeerById(peerID, testPeer)
	}

	senderPeerID := common.PeerId("sender-peer")

	// Call HandleAddr with addresses to forward
	addresses := []PeerAddress{
		{
			PeerId:              "discovered-peer-1",
			LastActiveTimestamp: now,
		},
		{
			PeerId:              "discovered-peer-2",
			LastActiveTimestamp: now,
		},
	}

	// Mock that the infrastructure already registered these peers as NOT connected
	peerStore.AddPeerById("discovered-peer-1", &peer.Peer{
		Version:  "1.0.0",
		State:    common.StateNew,
		LastSeen: now - 3600,
	})
	peerStore.AddPeerById("discovered-peer-2", &peer.Peer{
		Version:  "1.0.0",
		State:    common.StateNew,
		LastSeen: now - 3600,
	})

	service.HandleAddr(senderPeerID, addresses)

	// Verify that the sender never received the addresses
	sendAddrCalls := addrSender.sendAddrCalls
	for _, call := range sendAddrCalls {
		if call.peerID == senderPeerID {
			t.Errorf("expected addresses not to be forwarded to sender %s", senderPeerID)
		}
	}

	// Verify that other peers received the addresses
	forwardedTo := make(map[common.PeerId]bool)
	for _, call := range sendAddrCalls {
		forwardedTo[call.peerID] = true
	}

	// Sender should not be in forwardedTo
	if forwardedTo[senderPeerID] {
		t.Errorf("expected sender %s not to be in forwarded recipients", senderPeerID)
	}

	// Verify that sender's AddrsSentTo does NOT contain the discovered peer IDs
	senderPeer, _ := peerStore.GetPeer(senderPeerID)
	senderPeer.Lock()
	defer senderPeer.Unlock()
	for _, addr := range addresses {
		if senderPeer.AddrsSentTo.Contains(addr.PeerId) {
			t.Errorf("expected sender's AddrsSentTo not to contain %s", addr.PeerId)
		}
	}
}

// TestForwardAddrs_DoesNotForwardToPeersThatAlreadyReceived verifies that an address
// is not forwarded to a peer that has already received it.
func TestForwardAddrs_DoesNotForwardToPeersThatAlreadyReceived(t *testing.T) {
	peerStore := newMockDiscoveryPeerRetriever()
	addrSender := newMockAddrMsgSender()
	getAddrSender := newMockGetAddrMsgSender()
	peerCreator := newMockPeerCreator()

	service := NewDiscoveryService(nil, peerCreator, addrSender, peerStore, getAddrSender)

	// Create and add connected peers
	connectedPeers := []common.PeerId{"sender-peer", "peer-1", "peer-2", "peer-3"}
	now := time.Now().Unix()

	for _, peerID := range connectedPeers {
		testPeer := &peer.Peer{
			Version:     "1.0.0",
			State:       common.StateConnected,
			LastSeen:    now,
			AddrsSentTo: mapset.NewSet[common.PeerId](),
		}
		peerStore.AddPeerById(peerID, testPeer)
	}

	senderPeerID := common.PeerId("sender-peer")

	// Add the discovered peer to the store as NOT connected
	discoveredPeer := &peer.Peer{
		Version:  "1.0.0",
		State:    common.StateNew,
		LastSeen: now - 3600,
	}
	peerStore.AddPeerById("discovered-peer", discoveredPeer)

	// Mark peer-1 as already having received the discovered peer address
	peer1, _ := peerStore.GetPeer("peer-1")
	peer1.Lock()
	if peer1.AddrsSentTo == nil {
		peer1.AddrsSentTo = mapset.NewSet[common.PeerId]()
	}
	peer1.AddrsSentTo.Add(common.PeerId("discovered-peer"))
	peer1.Unlock()

	// Call HandleAddr with the address
	addresses := []PeerAddress{
		{
			PeerId:              "discovered-peer",
			LastActiveTimestamp: now,
		},
	}

	service.HandleAddr(senderPeerID, addresses)

	// Verify that peer-1 did not receive the address
	sendAddrCalls := addrSender.sendAddrCalls
	for _, call := range sendAddrCalls {
		if call.peerID == "peer-1" && len(call.addrs) > 0 && call.addrs[0].PeerId == "discovered-peer" {
			t.Errorf("expected address not to be forwarded to peer-1 which already received it")
		}
	}

	// Verify that peer-1 still has the discovered-peer in AddrsSentTo (it was set before the test)
	p1, _ := peerStore.GetPeer("peer-1")
	p1.Lock()
	defer p1.Unlock()
	if !p1.AddrsSentTo.Contains(common.PeerId("discovered-peer")) {
		t.Errorf("expected peer-1's AddrsSentTo to still contain discovered-peer")
	}

	// Verify that peers who DID receive the address have their AddrsSentTo updated
	for _, call := range sendAddrCalls {
		if len(call.addrs) > 0 && call.addrs[0].PeerId == "discovered-peer" {
			recipient, _ := peerStore.GetPeer(call.peerID)
			recipient.Lock()
			if !recipient.AddrsSentTo.Contains(common.PeerId("discovered-peer")) {
				recipient.Unlock()
				t.Errorf("expected recipient %s to have discovered-peer in AddrsSentTo", call.peerID)
			} else {
				recipient.Unlock()
			}
		}
	}
}

// TestForwardAddrs_ForwardsToRandomPeers verifies that for each address,
// 2 random peers (or fewer if not enough connected peers) are selected to forward to.
func TestForwardAddrs_ForwardsToRandomPeers(t *testing.T) {
	peerStore := newMockDiscoveryPeerRetriever()
	addrSender := newMockAddrMsgSender()
	getAddrSender := newMockGetAddrMsgSender()
	peerCreator := newMockPeerCreator()

	service := NewDiscoveryService(nil, peerCreator, addrSender, peerStore, getAddrSender)

	// Create and add connected peers
	connectedPeers := []common.PeerId{"sender-peer", "peer-1", "peer-2", "peer-3", "peer-4"}
	now := time.Now().Unix()

	for _, peerID := range connectedPeers {
		testPeer := &peer.Peer{
			Version:     "1.0.0",
			State:       common.StateConnected,
			LastSeen:    now,
			AddrsSentTo: mapset.NewSet[common.PeerId](),
		}
		peerStore.AddPeerById(peerID, testPeer)
	}

	senderPeerID := common.PeerId("sender-peer")

	// Add discovered peers to the store as NOT connected
	discoveredPeers := []common.PeerId{"discovered-1", "discovered-2", "discovered-3"}
	for _, peerID := range discoveredPeers {
		discoveredPeer := &peer.Peer{
			Version:  "1.0.0",
			State:    common.StateNew,
			LastSeen: now - 3600,
		}
		peerStore.AddPeerById(peerID, discoveredPeer)
	}

	// Call HandleAddr with multiple addresses
	addresses := []PeerAddress{
		{
			PeerId:              "discovered-1",
			LastActiveTimestamp: now,
		},
		{
			PeerId:              "discovered-2",
			LastActiveTimestamp: now,
		},
		{
			PeerId:              "discovered-3",
			LastActiveTimestamp: now,
		},
	}

	service.HandleAddr(senderPeerID, addresses)

	// Each address should be forwarded to exactly 2 peers
	// Collect forwarding counts per address
	forwardCountPerAddr := make(map[common.PeerId]int)
	sendAddrCalls := addrSender.sendAddrCalls

	for _, call := range sendAddrCalls {
		for _, addr := range call.addrs {
			forwardCountPerAddr[addr.PeerId]++
		}
	}

	// Each discovered peer should have exactly 2 forwards
	for _, discoveredID := range discoveredPeers {
		count, exists := forwardCountPerAddr[discoveredID]
		if !exists {
			t.Errorf("expected address %s to be forwarded", discoveredID)
		}
		if count != 2 {
			t.Errorf("expected address %s to be forwarded to exactly 2 peers, got %d", discoveredID, count)
		}
	}

	// Verify that peers who received each address have their AddrsSentTo updated
	for _, call := range sendAddrCalls {
		for _, addr := range call.addrs {
			recipient, _ := peerStore.GetPeer(call.peerID)
			recipient.Lock()
			if !recipient.AddrsSentTo.Contains(addr.PeerId) {
				recipient.Unlock()
				t.Errorf("expected recipient %s to have %s in AddrsSentTo", call.peerID, addr.PeerId)
			} else {
				recipient.Unlock()
			}
		}
	}
}

// TestForwardAddrs_WithLimitedPeers verifies that when there are fewer than 2
// eligible peers (after excluding sender), all available peers are used for forwarding.
func TestForwardAddrs_WithLimitedPeers(t *testing.T) {
	peerStore := newMockDiscoveryPeerRetriever()
	addrSender := newMockAddrMsgSender()
	getAddrSender := newMockGetAddrMsgSender()
	peerCreator := newMockPeerCreator()

	service := NewDiscoveryService(nil, peerCreator, addrSender, peerStore, getAddrSender)

	// Create only sender and 1 other connected peer
	connectedPeers := []common.PeerId{"sender-peer", "peer-1"}
	now := time.Now().Unix()

	for _, peerID := range connectedPeers {
		testPeer := &peer.Peer{
			Version:     "1.0.0",
			State:       common.StateConnected,
			LastSeen:    now,
			AddrsSentTo: mapset.NewSet[common.PeerId](),
		}
		peerStore.AddPeerById(peerID, testPeer)
	}

	senderPeerID := common.PeerId("sender-peer")

	// Add the discovered peer as NOT connected (it's being discovered, not connected yet)
	discoveredPeer := &peer.Peer{
		Version:  "1.0.0",
		State:    common.StateNew,
		LastSeen: now - 3600,
	}
	peerStore.AddPeerById("discovered-peer", discoveredPeer)

	// Call HandleAddr with the address
	addresses := []PeerAddress{
		{
			PeerId:              "discovered-peer",
			LastActiveTimestamp: now,
		},
	}

	service.HandleAddr(senderPeerID, addresses)

	// The address should be forwarded to exactly 1 peer (peer-1)
	sendAddrCalls := addrSender.sendAddrCalls
	forwardCount := 0
	var recipientPeerID common.PeerId
	for _, call := range sendAddrCalls {
		for _, addr := range call.addrs {
			if addr.PeerId == "discovered-peer" {
				forwardCount++
				recipientPeerID = call.peerID
			}
		}
	}

	if forwardCount != 1 {
		t.Errorf("expected address to be forwarded to exactly 1 peer when only 1 eligible peer exists, got %d", forwardCount)
	}

	// Verify that the recipient's AddrsSentTo is updated
	if recipientPeerID != "" {
		recipient, _ := peerStore.GetPeer(recipientPeerID)
		recipient.Lock()
		if !recipient.AddrsSentTo.Contains(common.PeerId("discovered-peer")) {
			recipient.Unlock()
			t.Errorf("expected recipient %s to have discovered-peer in AddrsSentTo", recipientPeerID)
		} else {
			recipient.Unlock()
		}
	}
}

// TestForwardAddrs_WithNoEligiblePeers verifies that when there are no
// eligible peers (only sender exists), no forwarding occurs.
func TestForwardAddrs_WithNoEligiblePeers(t *testing.T) {
	peerStore := newMockDiscoveryPeerRetriever()
	addrSender := newMockAddrMsgSender()
	getAddrSender := newMockGetAddrMsgSender()
	peerCreator := newMockPeerCreator()

	service := NewDiscoveryService(nil, peerCreator, addrSender, peerStore, getAddrSender)

	// Create only the sender peer
	now := time.Now().Unix()
	senderPeer := &peer.Peer{
		Version:     "1.0.0",
		State:       common.StateConnected,
		LastSeen:    now,
		AddrsSentTo: mapset.NewSet[common.PeerId](),
	}
	peerStore.AddPeerById("sender-peer", senderPeer)

	senderPeerID := common.PeerId("sender-peer")

	// Add the discovered peer as NOT connected
	discoveredPeer := &peer.Peer{
		Version:  "1.0.0",
		State:    common.StateNew,
		LastSeen: now - 3600,
	}
	peerStore.AddPeerById("discovered-peer", discoveredPeer)

	// Call HandleAddr with the address
	addresses := []PeerAddress{
		{
			PeerId:              "discovered-peer",
			LastActiveTimestamp: now,
		},
	}

	service.HandleAddr(senderPeerID, addresses)

	// No forwarding should occur
	sendAddrCalls := addrSender.sendAddrCalls
	if len(sendAddrCalls) != 0 {
		t.Errorf("expected no forwarding when no eligible peers exist, got %d calls", len(sendAddrCalls))
	}

	// Verify that no peers have the discovered-peer in AddrsSentTo (except the discovered peer itself)
	allPeers := peerStore.GetAllConnectedPeers()
	for _, peerID := range allPeers {
		if peerID == "discovered-peer" {
			continue // Skip the discovered peer itself
		}
		p, _ := peerStore.GetPeer(peerID)
		p.Lock()
		if p.AddrsSentTo.Contains(common.PeerId("discovered-peer")) {
			p.Unlock()
			t.Errorf("expected peer %s not to have discovered-peer in AddrsSentTo when no forwarding occurred", peerID)
		} else {
			p.Unlock()
		}
	}
}

// TestForwardAddrs_IndependentPeerSelection verifies that each address
// independently selects 2 random peers.
func TestForwardAddrs_IndependentPeerSelection(t *testing.T) {
	peerStore := newMockDiscoveryPeerRetriever()
	addrSender := newMockAddrMsgSender()
	getAddrSender := newMockGetAddrMsgSender()
	peerCreator := newMockPeerCreator()

	service := NewDiscoveryService(nil, peerCreator, addrSender, peerStore, getAddrSender)

	// Create multiple connected peers
	connectedPeers := []common.PeerId{"sender-peer", "peer-1", "peer-2", "peer-3", "peer-4"}
	now := time.Now().Unix()

	for _, peerID := range connectedPeers {
		testPeer := &peer.Peer{
			Version:     "1.0.0",
			State:       common.StateConnected,
			LastSeen:    now,
			AddrsSentTo: mapset.NewSet[common.PeerId](),
		}
		peerStore.AddPeerById(peerID, testPeer)
	}

	senderPeerID := common.PeerId("sender-peer")

	// Add multiple discovered peers as NOT connected
	discoveredPeers := []common.PeerId{"discovered-1", "discovered-2", "discovered-3", "discovered-4", "discovered-5"}
	for _, peerID := range discoveredPeers {
		discoveredPeer := &peer.Peer{
			Version:  "1.0.0",
			State:    common.StateNew,
			LastSeen: now - 3600,
		}
		peerStore.AddPeerById(peerID, discoveredPeer)
	}

	// Call HandleAddr with multiple addresses
	addresses := make([]PeerAddress, 0, len(discoveredPeers))
	for _, peerID := range discoveredPeers {
		addresses = append(addresses, PeerAddress{
			PeerId:              peerID,
			LastActiveTimestamp: now,
		})
	}

	service.HandleAddr(senderPeerID, addresses)

	// Verify that different addresses can be forwarded to different peers
	// Collect which peers received which addresses
	addrToRecipients := make(map[common.PeerId][]common.PeerId)
	sendAddrCalls := addrSender.sendAddrCalls

	for _, call := range sendAddrCalls {
		for _, addr := range call.addrs {
			addrToRecipients[addr.PeerId] = append(addrToRecipients[addr.PeerId], call.peerID)
		}
	}

	// Each address should have exactly 2 recipients
	for _, addr := range addresses {
		recipients := addrToRecipients[addr.PeerId]
		if len(recipients) != 2 {
			t.Errorf("expected address %s to have exactly 2 recipients, got %d", addr.PeerId, len(recipients))
		}
	}

	// Verify that all recipients have AddrsSentTo updated
	for _, call := range sendAddrCalls {
		for _, addr := range call.addrs {
			recipient, _ := peerStore.GetPeer(call.peerID)
			recipient.Lock()
			if !recipient.AddrsSentTo.Contains(addr.PeerId) {
				recipient.Unlock()
				t.Errorf("expected recipient %s to have %s in AddrsSentTo", call.peerID, addr.PeerId)
			} else {
				recipient.Unlock()
			}
		}
	}
}

// Test for local peer exclusion from timestamp updating

// TestHandleAddr_ExcludesLocalPeerFromTimestampUpdate verifies that when
// the local peer's address is received, it is excluded from timestamp updating.
func TestHandleAddr_ExcludesLocalPeerFromTimestampUpdate(t *testing.T) {
	peerStore := newMockDiscoveryPeerRetriever()
	addrSender := newMockAddrMsgSender()
	getAddrSender := newMockGetAddrMsgSender()
	peerCreator := newMockPeerCreator()

	service := NewDiscoveryService(nil, peerCreator, addrSender, peerStore, getAddrSender)

	// Create and add the local peer with an old timestamp
	oldTimestamp := time.Now().Add(-24 * time.Hour).Unix()
	localPeer := &peer.Peer{
		Version:     "1.0.0",
		State:       common.StateConnected,
		LastSeen:    oldTimestamp,
		AddrsSentTo: mapset.NewSet[common.PeerId](),
	}
	peerStore.AddPeerById("local-peer", localPeer)

	senderPeerID := common.PeerId("sender-peer")
	newTimestamp := time.Now().Unix()

	// Call HandleAddr with local peer's address (with a newer timestamp)
	addresses := []PeerAddress{
		{
			PeerId:              "local-peer",
			LastActiveTimestamp: newTimestamp,
		},
	}

	service.HandleAddr(senderPeerID, addresses)

	// Verify the local peer's LastSeen was NOT updated
	retrievedPeer, exists := peerStore.GetPeer("local-peer")
	if !exists {
		t.Fatal("expected local peer to exist")
	}

	if retrievedPeer.LastSeen != oldTimestamp {
		t.Errorf("expected local peer LastSeen to remain %d (not updated to %d)", oldTimestamp, newTimestamp)
	}
}

// TestHandleAddr_ExcludesLocalPeerFromForwarding verifies that when ONLY the local
// peer's address is received, it is excluded from both timestamp updating and forwarding.
// When the local peer's address is received along with other addresses, it IS forwarded.
func TestHandleAddr_ExcludesLocalPeerFromForwarding(t *testing.T) {
	peerStore := newMockDiscoveryPeerRetriever()
	addrSender := newMockAddrMsgSender()
	getAddrSender := newMockGetAddrMsgSender()
	peerCreator := newMockPeerCreator()

	service := NewDiscoveryService(nil, peerCreator, addrSender, peerStore, getAddrSender)

	// Create and add connected peers
	connectedPeers := []common.PeerId{"sender-peer", "local-peer", "peer-1", "peer-2"}
	now := time.Now().Unix()

	for _, peerID := range connectedPeers {
		testPeer := &peer.Peer{
			Version:     "1.0.0",
			State:       common.StateConnected,
			LastSeen:    now,
			AddrsSentTo: mapset.NewSet[common.PeerId](),
		}
		peerStore.AddPeerById(peerID, testPeer)
	}

	senderPeerID := common.PeerId("sender-peer")
	newTimestamp := time.Now().Unix()

	// Call HandleAddr with ONLY local peer's address (with a newer timestamp)
	addresses := []PeerAddress{
		{
			PeerId:              "local-peer",
			LastActiveTimestamp: newTimestamp,
		},
	}

	service.HandleAddr(senderPeerID, addresses)

	// Verify the local peer's LastSeen was NOT updated
	localPeer, exists := peerStore.GetPeer("local-peer")
	if !exists {
		t.Fatal("expected local peer to exist")
	}

	if localPeer.LastSeen != now {
		t.Errorf("expected local peer LastSeen to remain unchanged")
	}

	// When only the local peer address is received, nothing is forwarded
	// Verify that the local peer's address was NOT forwarded
	sendAddrCalls := addrSender.sendAddrCalls
	forwardedLocalPeer := false
	for _, call := range sendAddrCalls {
		for _, addr := range call.addrs {
			if addr.PeerId == "local-peer" {
				forwardedLocalPeer = true
				break
			}
		}
		if forwardedLocalPeer {
			break
		}
	}

	if forwardedLocalPeer {
		t.Errorf("expected local peer's address NOT to be forwarded when it's the only address")
	}

	// Verify that no peer has the local peer's address in AddrsSentTo
	allPeers := peerStore.GetAllConnectedPeers()
	for _, peerID := range allPeers {
		p, _ := peerStore.GetPeer(peerID)
		p.Lock()
		if p.AddrsSentTo.Contains(common.PeerId("local-peer")) {
			p.Unlock()
			t.Errorf("expected peer %s not to have local-peer in AddrsSentTo when it's the only address", peerID)
		} else {
			p.Unlock()
		}
	}
}

// TestHandleAddr_MultipleAddressesIncludingLocalPeer verifies that when
// multiple addresses are received including the local peer's address,
// only the local peer is excluded from timestamp updating.
func TestHandleAddr_MultipleAddressesIncludingLocalPeer(t *testing.T) {
	peerStore := newMockDiscoveryPeerRetriever()
	addrSender := newMockAddrMsgSender()
	getAddrSender := newMockGetAddrMsgSender()
	peerCreator := newMockPeerCreator()

	service := NewDiscoveryService(nil, peerCreator, addrSender, peerStore, getAddrSender)

	// Create and add multiple peers with old timestamps
	oldTimestamp := time.Now().Add(-24 * time.Hour).Unix()
	localPeer := &peer.Peer{
		Version:     "1.0.0",
		State:       common.StateConnected,
		LastSeen:    oldTimestamp,
		AddrsSentTo: mapset.NewSet[common.PeerId](),
	}
	peerStore.AddPeerById("local-peer", localPeer)

	peer1 := &peer.Peer{
		Version:     "1.0.0",
		State:       common.StateConnected,
		LastSeen:    oldTimestamp,
		AddrsSentTo: mapset.NewSet[common.PeerId](),
	}
	peerStore.AddPeerById("peer-1", peer1)

	peer2 := &peer.Peer{
		Version:     "1.0.0",
		State:       common.StateConnected,
		LastSeen:    oldTimestamp,
		AddrsSentTo: mapset.NewSet[common.PeerId](),
	}
	peerStore.AddPeerById("peer-2", peer2)

	senderPeerID := common.PeerId("sender-peer")
	newTimestamp := time.Now().Unix()

	// Call HandleAddr with multiple addresses including local peer
	addresses := []PeerAddress{
		{
			PeerId:              "local-peer",
			LastActiveTimestamp: newTimestamp,
		},
		{
			PeerId:              "peer-1",
			LastActiveTimestamp: newTimestamp,
		},
		{
			PeerId:              "peer-2",
			LastActiveTimestamp: newTimestamp,
		},
	}

	service.HandleAddr(senderPeerID, addresses)

	// Verify local peer's LastSeen was NOT updated
	retrievedLocalPeer, _ := peerStore.GetPeer("local-peer")
	if retrievedLocalPeer.LastSeen != oldTimestamp {
		t.Errorf("expected local peer LastSeen to remain %d, got %d", oldTimestamp, retrievedLocalPeer.LastSeen)
	}

	// Verify other peers' LastSeen were updated
	retrievedPeer1, _ := peerStore.GetPeer("peer-1")
	if retrievedPeer1.LastSeen != newTimestamp {
		t.Errorf("expected peer-1 LastSeen to be updated to %d, got %d", newTimestamp, retrievedPeer1.LastSeen)
	}

	retrievedPeer2, _ := peerStore.GetPeer("peer-2")
	if retrievedPeer2.LastSeen != newTimestamp {
		t.Errorf("expected peer-2 LastSeen to be updated to %d, got %d", newTimestamp, retrievedPeer2.LastSeen)
	}

	// Verify that peers who received addresses have their AddrsSentTo updated
	// By design, ALL addresses including local peer's address are forwarded
	sendAddrCalls := addrSender.sendAddrCalls
	for _, call := range sendAddrCalls {
		for _, addr := range call.addrs {
			recipient, _ := peerStore.GetPeer(call.peerID)
			recipient.Lock()
			if !recipient.AddrsSentTo.Contains(addr.PeerId) {
				recipient.Unlock()
				t.Errorf("expected recipient %s to have %s in AddrsSentTo", call.peerID, addr.PeerId)
			} else {
				recipient.Unlock()
			}
		}
	}
}
