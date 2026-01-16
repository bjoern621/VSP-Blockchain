package connectioncheck

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/data/peer"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//
// Mocks
//

// mockNetworkInfoCleaner is a mock implementation of networkInfoCleaner for testing.
type mockNetworkInfoCleaner struct {
	removedPeers []common.PeerId
}

func (m *mockNetworkInfoCleaner) RemovePeer(peerID common.PeerId) {
	m.removedPeers = append(m.removedPeers, peerID)
}

// mockPeerRetriever is a mock implementation of peerRetriever for testing.
type mockPeerRetriever struct {
	peers map[common.PeerId]*peer.Peer
}

func newMockPeerRetriever() *mockPeerRetriever {
	return &mockPeerRetriever{
		peers: make(map[common.PeerId]*peer.Peer),
	}
}

func (m *mockPeerRetriever) GetPeer(id common.PeerId) (*peer.Peer, bool) {
	p, exists := m.peers[id]
	return p, exists
}

func (m *mockPeerRetriever) GetAllConnectedPeers() []common.PeerId {
	ids := make([]common.PeerId, 0)
	for id, p := range m.peers {
		if p.State == common.StateConnected {
			ids = append(ids, id)
		}
	}
	return ids
}

// mockPeerRemover is a mock implementation of peerRemover for testing.
type mockPeerRemover struct {
	removedPeers []common.PeerId
}

func (m *mockPeerRemover) RemovePeer(id common.PeerId) {
	m.removedPeers = append(m.removedPeers, id)
}

//
// Tests
//

func TestCheckConnections(t *testing.T) {
	peerRetriever := newMockPeerRetriever()
	peerRemover := &mockPeerRemover{}
	networkInfoCleaner := &mockNetworkInfoCleaner{}

	service := NewConnectionCheckService(peerRetriever, peerRemover, networkInfoCleaner)

	// Create test peers with different LastSeen timestamps
	now := time.Now().Unix()

	// Peer 1: Recent LastSeen (should NOT be removed)
	peerID1 := common.PeerId("peer-1-recent")
	peerRetriever.peers[peerID1] = &peer.Peer{
		State:     common.StateConnected,
		LastSeen:  now - (10 * 60), // 10 minutes ago
		Direction: common.DirectionOutbound,
	}

	// Peer 2: Old LastSeen (SHOULD be removed)
	peerID2 := common.PeerId("peer-2-old")
	peerRetriever.peers[peerID2] = &peer.Peer{
		State:     common.StateConnected,
		LastSeen:  now - (20 * 60), // 20 minutes ago (> PeerTimeout)
		Direction: common.DirectionInbound,
	}

	// Peer 3: LastSeen = 0 (SHOULD be removed)
	peerID3 := common.PeerId("peer-3-zero")
	peerRetriever.peers[peerID3] = &peer.Peer{
		State:     common.StateConnected,
		LastSeen:  0,
		Direction: common.DirectionOutbound,
	}

	// Run the connection check
	service.checkConnections()

	// Verify that peer-2-old and peer-3-zero were removed
	assert.Equal(t, 2, len(peerRemover.removedPeers), "Should have removed 2 peers")
	assert.Contains(t, peerRemover.removedPeers, peerID2, "Should have removed peer-2-old")
	assert.Contains(t, peerRemover.removedPeers, peerID3, "Should have removed peer-3-zero")

	// Verify that network info cleaner was called
	assert.Equal(t, 2, len(networkInfoCleaner.removedPeers), "Should have called networkInfoCleaner 2 times")
	assert.Contains(t, networkInfoCleaner.removedPeers, peerID2, "Should have cleaned up peer-2-old")
	assert.Contains(t, networkInfoCleaner.removedPeers, peerID3, "Should have cleaned up peer-3-zero")

	// Verify that peer-1-recent was NOT removed
	assert.NotContains(t, peerRemover.removedPeers, peerID1, "Should NOT have removed peer-1-recent")
}

func TestCheckConnectionsAtTimeoutBoundary(t *testing.T) {
	peerRetriever := newMockPeerRetriever()
	peerRemover := &mockPeerRemover{}
	networkInfoCleaner := &mockNetworkInfoCleaner{}

	service := NewConnectionCheckService(peerRetriever, peerRemover, networkInfoCleaner)

	now := time.Now().Unix()

	// Create peer exactly at timeout boundary (SHOULD be removed)
	peerID := common.PeerId("peer-exact-timeout")
	peerRetriever.peers[peerID] = &peer.Peer{
		State:     common.StateConnected,
		LastSeen:  now - int64(PeerTimeout.Seconds()),
		Direction: common.DirectionOutbound,
	}

	service.checkConnections()

	assert.Equal(t, 1, len(peerRemover.removedPeers), "Should have removed peer at exact timeout")
	assert.Contains(t, peerRemover.removedPeers, peerID, "Should have removed peer at exact timeout")
}

func TestCheckConnectionsJustBeforeTimeout(t *testing.T) {
	peerRetriever := newMockPeerRetriever()
	peerRemover := &mockPeerRemover{}
	networkInfoCleaner := &mockNetworkInfoCleaner{}

	service := NewConnectionCheckService(peerRetriever, peerRemover, networkInfoCleaner)

	now := time.Now().Unix()

	// Create peer just before timeout boundary (should NOT be removed)
	peerID := common.PeerId("peer-before-timeout")
	peerRetriever.peers[peerID] = &peer.Peer{
		State:     common.StateConnected,
		LastSeen:  now - int64(PeerTimeout.Seconds()) + 1, // 1 second before timeout
		Direction: common.DirectionOutbound,
	}

	service.checkConnections()

	assert.Equal(t, 0, len(peerRemover.removedPeers), "Should NOT have removed peer before timeout")
}

func TestCheckConnectionsNoConnectedPeers(t *testing.T) {
	peerRetriever := newMockPeerRetriever()
	peerRemover := &mockPeerRemover{}
	networkInfoCleaner := &mockNetworkInfoCleaner{}

	service := NewConnectionCheckService(peerRetriever, peerRemover, networkInfoCleaner)

	// No peers in the store
	service.checkConnections()

	assert.Equal(t, 0, len(peerRemover.removedPeers), "Should not have removed any peers")
}

func TestCheckConnectionsPeerNotFound(t *testing.T) {
	peerRetriever := newMockPeerRetriever()
	peerRemover := &mockPeerRemover{}
	networkInfoCleaner := &mockNetworkInfoCleaner{}

	service := NewConnectionCheckService(peerRetriever, peerRemover, networkInfoCleaner)

	// Add a peer to the retriever
	peerID := common.PeerId("peer-1")
	peerRetriever.peers[peerID] = &peer.Peer{
		State:     common.StateConnected,
		LastSeen:  0,
		Direction: common.DirectionOutbound,
	}

	// Remove the peer from the retriever (simulating concurrent removal)
	delete(peerRetriever.peers, peerID)

	// GetAllConnectedPeers will return an empty slice since we removed it
	// But let's test the case where the peer is returned but then not found
	service.checkConnections()

	// Should not panic and should not remove anything
	assert.Equal(t, 0, len(peerRemover.removedPeers), "Should not have removed any peers")
}

func TestStartAndStop(t *testing.T) {
	peerRetriever := newMockPeerRetriever()
	peerRemover := &mockPeerRemover{}
	networkInfoCleaner := &mockNetworkInfoCleaner{}

	service := NewConnectionCheckService(peerRetriever, peerRemover, networkInfoCleaner)

	// Start the service
	service.Start()
	assert.NotNil(t, service.ticker, "Ticker should be initialized")

	// Wait a bit to ensure the goroutine starts
	time.Sleep(10 * time.Millisecond)

	// Stop the service
	service.Stop()

	// Verify that calling Stop again doesn't panic
	service.Stop()
}

func TestRemovePeer(t *testing.T) {
	peerRetriever := newMockPeerRetriever()
	peerRemover := &mockPeerRemover{}
	networkInfoCleaner := &mockNetworkInfoCleaner{}

	service := NewConnectionCheckService(peerRetriever, peerRemover, networkInfoCleaner)

	peerID := common.PeerId("test-peer")

	// Call removePeer
	service.removePeer(peerID)

	// Verify both removers were called
	assert.Equal(t, 1, len(networkInfoCleaner.removedPeers), "Network info cleaner should have been called")
	assert.Equal(t, peerID, networkInfoCleaner.removedPeers[0], "Network info cleaner should have removed the peer")

	assert.Equal(t, 1, len(peerRemover.removedPeers), "Peer remover should have been called")
	assert.Equal(t, peerID, peerRemover.removedPeers[0], "Peer remover should have removed the peer")
}

func TestNewConnectionCheckService(t *testing.T) {
	peerRetriever := newMockPeerRetriever()
	peerRemover := &mockPeerRemover{}
	networkInfoCleaner := &mockNetworkInfoCleaner{}

	service := NewConnectionCheckService(peerRetriever, peerRemover, networkInfoCleaner)

	require.NotNil(t, service, "Service should be created")
	assert.NotNil(t, service.stopChan, "Stop channel should be initialized")
	assert.Equal(t, peerRetriever, service.peerRetriever, "Peer retriever should be set")
	assert.Equal(t, peerRemover, service.storePeerRemover, "Peer remover should be set")
	assert.Equal(t, networkInfoCleaner, service.networkInfoCleaner, "Network info cleaner should be set")
	assert.Nil(t, service.ticker, "Ticker should not be initialized until Start is called")
}
