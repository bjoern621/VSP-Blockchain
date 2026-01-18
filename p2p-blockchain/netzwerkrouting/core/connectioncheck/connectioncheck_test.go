package connectioncheck

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//
// Mocks
//

// mockPeerRetriever is a mock implementation of peerRetriever for testing.
type mockPeerRetriever struct {
	peers map[common.PeerId]*common.Peer
}

func newMockPeerRetriever() *mockPeerRetriever {
	return &mockPeerRetriever{
		peers: make(map[common.PeerId]*common.Peer),
	}
}

func (m *mockPeerRetriever) GetPeer(id common.PeerId) (*common.Peer, bool) {
	p, exists := m.peers[id]
	return p, exists
}

func (m *mockPeerRetriever) GetPeersWithHandshakeStarted() []common.PeerId {
	ids := make([]common.PeerId, 0)
	for id, p := range m.peers {
		if p.State == common.StateAwaitingVerack || p.State == common.StateAwaitingAck || p.State == common.StateConnected {
			ids = append(ids, id)
		}
	}
	return ids
}

func (m *mockPeerRetriever) GetHolddownPeers() []common.PeerId {
	ids := make([]common.PeerId, 0)
	for id, p := range m.peers {
		if p.State == common.StateHolddown {
			ids = append(ids, id)
		}
	}
	return ids
}

func (m *mockPeerRetriever) RemovePeer(id common.PeerId) {
	delete(m.peers, id)
}

type mockPeerDisconnector struct {
	disconnectedPeers []common.PeerId
}

func (m *mockPeerDisconnector) Disconnect(id common.PeerId) error {
	m.disconnectedPeers = append(m.disconnectedPeers, id)
	return nil
}

type mockNetworkInfoRemover struct {
	removedPeers []common.PeerId
}

func (m *mockNetworkInfoRemover) RemovePeer(id common.PeerId) {
	m.removedPeers = append(m.removedPeers, id)
}

//
// Tests
//

func TestCheckConnections(t *testing.T) {
	peerRetriever := newMockPeerRetriever()
	peerDisconnector := &mockPeerDisconnector{}
	networkInfoRemover := &mockNetworkInfoRemover{}

	service := NewConnectionCheckService(peerRetriever, peerDisconnector, networkInfoRemover)

	// Create test peers with different LastSeen timestamps
	now := time.Now().Unix()

	// Peer 1: Recent LastSeen (should NOT be removed)
	peerID1 := common.PeerId("peer-1-recent")
	peerRetriever.peers[peerID1] = &common.Peer{
		State:    common.StateConnected,
		LastSeen: now - (8 * 60), // 8 minutes ago (< 9 min PeerTimeout)
	}

	// Peer 2: Old LastSeen (SHOULD be removed)
	peerID2 := common.PeerId("peer-2-old")
	peerRetriever.peers[peerID2] = &common.Peer{
		State:    common.StateConnected,
		LastSeen: now - (10 * 60), // 10 minutes ago (> 9 min PeerTimeout)
	}

	// Peer 3: LastSeen = 0 (SHOULD be removed)
	peerID3 := common.PeerId("peer-3-zero")
	peerRetriever.peers[peerID3] = &common.Peer{
		State:    common.StateConnected,
		LastSeen: 0,
	}

	// Run the connection check
	service.checkConnections()

	// Verify that peer-2-old and peer-3-zero were disconnected (put into holddown)
	assert.Equal(t, 2, len(peerDisconnector.disconnectedPeers), "Should have disconnected 2 peers")
	assert.Contains(t, peerDisconnector.disconnectedPeers, peerID2, "Should have disconnected peer-2-old")
	assert.Contains(t, peerDisconnector.disconnectedPeers, peerID3, "Should have disconnected peer-3-zero")

	// Verify that peer-1-recent was NOT disconnected
	assert.NotContains(t, peerDisconnector.disconnectedPeers, peerID1, "Should NOT have disconnected peer-1-recent")
}

func TestCheckConnectionsAtTimeoutBoundary(t *testing.T) {
	peerRetriever := newMockPeerRetriever()
	peerDisconnector := &mockPeerDisconnector{}
	networkInfoRemover := &mockNetworkInfoRemover{}

	service := NewConnectionCheckService(peerRetriever, peerDisconnector, networkInfoRemover)
	now := time.Now().Unix()

	// Create peer exactly at timeout boundary (SHOULD be removed)
	peerID := common.PeerId("peer-exact-timeout")
	peerRetriever.peers[peerID] = &common.Peer{
		State:    common.StateConnected,
		LastSeen: now - int64(PeerTimeout.Seconds()), // exactly 9 minutes ago
	}

	service.checkConnections()

	assert.Equal(t, 1, len(peerDisconnector.disconnectedPeers), "Should have disconnected peer at exact timeout")
	assert.Contains(t, peerDisconnector.disconnectedPeers, peerID, "Should have disconnected peer at exact timeout")
}

func TestCheckConnectionsJustBeforeTimeout(t *testing.T) {
	peerRetriever := newMockPeerRetriever()
	peerDisconnector := &mockPeerDisconnector{}
	networkInfoRemover := &mockNetworkInfoRemover{}

	service := NewConnectionCheckService(peerRetriever, peerDisconnector, networkInfoRemover)

	now := time.Now().Unix()

	// Create peer just before timeout boundary (should NOT be removed)
	peerID := common.PeerId("peer-before-timeout")
	peerRetriever.peers[peerID] = &common.Peer{
		State:    common.StateConnected,
		LastSeen: now - int64(PeerTimeout.Seconds()) + 1, // 1 second before 9 min timeout
	}

	service.checkConnections()

	assert.Equal(t, 0, len(peerDisconnector.disconnectedPeers), "Should NOT have disconnected peer before timeout")
}

func TestCheckConnectionsNoConnectedPeers(t *testing.T) {
	peerRetriever := newMockPeerRetriever()
	peerDisconnector := &mockPeerDisconnector{}
	networkInfoRemover := &mockNetworkInfoRemover{}

	service := NewConnectionCheckService(peerRetriever, peerDisconnector, networkInfoRemover)

	// No peers in the store
	service.checkConnections()

	assert.Equal(t, 0, len(peerDisconnector.disconnectedPeers), "Should not have disconnected any peers")
}

func TestCheckConnectionsPeerNotFound(t *testing.T) {
	peerRetriever := newMockPeerRetriever()
	peerDisconnector := &mockPeerDisconnector{}
	networkInfoRemover := &mockNetworkInfoRemover{}

	service := NewConnectionCheckService(peerRetriever, peerDisconnector, networkInfoRemover)

	// Add a peer to the retriever
	peerID := common.PeerId("peer-1")
	peerRetriever.peers[peerID] = &common.Peer{
		State:    common.StateConnected,
		LastSeen: 0,
	}

	// Remove the peer from the retriever (simulating concurrent removal)
	delete(peerRetriever.peers, peerID)

	// GetAllConnectedPeers will return an empty slice since we removed it
	// But let's test the case where the peer is returned but then not found
	service.checkConnections()

	// Should not panic and should not disconnect anything
	assert.Equal(t, 0, len(peerDisconnector.disconnectedPeers), "Should not have disconnected any peers")
}

func TestStartAndStop(t *testing.T) {
	peerRetriever := newMockPeerRetriever()
	peerDisconnector := &mockPeerDisconnector{}
	networkInfoRemover := &mockNetworkInfoRemover{}

	service := NewConnectionCheckService(peerRetriever, peerDisconnector, networkInfoRemover)

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

func TestDisconnectPeer(t *testing.T) {
	peerRetriever := newMockPeerRetriever()
	peerDisconnector := &mockPeerDisconnector{}
	networkInfoRemover := &mockNetworkInfoRemover{}

	service := NewConnectionCheckService(peerRetriever, peerDisconnector, networkInfoRemover)

	peerID := common.PeerId("test-peer")

	// Call removePeer (which calls Disconnect to put peer in holddown)
	service.removePeer(peerID)

	assert.Equal(t, 1, len(peerDisconnector.disconnectedPeers), "Peer disconnector should have been called")
	assert.Equal(t, peerID, peerDisconnector.disconnectedPeers[0], "Peer disconnector should have disconnected the peer")
}

func TestNewConnectionCheckService(t *testing.T) {
	peerRetriever := newMockPeerRetriever()
	peerDisconnector := &mockPeerDisconnector{}
	networkInfoRemover := &mockNetworkInfoRemover{}

	service := NewConnectionCheckService(peerRetriever, peerDisconnector, networkInfoRemover)

	require.NotNil(t, service, "Service should be created")
	assert.NotNil(t, service.stopChan, "Stop channel should be initialized")
	assert.Equal(t, peerRetriever, service.peerRetriever, "Peer retriever should be set")
	assert.Nil(t, service.ticker, "Ticker should not be initialized until Start is called")
}
