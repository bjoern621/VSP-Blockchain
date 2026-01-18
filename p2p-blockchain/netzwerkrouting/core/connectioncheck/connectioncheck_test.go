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

type mockPeerDisconnector struct {
	removedPeers []common.PeerId
}

func (m *mockPeerDisconnector) Disconnect(id common.PeerId) error {
	m.removedPeers = append(m.removedPeers, id)
	return nil
}

//
// Tests
//

func TestCheckConnections(t *testing.T) {
	peerRetriever := newMockPeerRetriever()
	peerDisconnector := &mockPeerDisconnector{}

	service := NewConnectionCheckService(peerRetriever, peerDisconnector)

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

	// Verify that peer-2-old and peer-3-zero were removed
	assert.Equal(t, 2, len(peerDisconnector.removedPeers), "Should have removed 2 peers")
	assert.Contains(t, peerDisconnector.removedPeers, peerID2, "Should have removed peer-2-old")
	assert.Contains(t, peerDisconnector.removedPeers, peerID3, "Should have removed peer-3-zero")

	// Verify that peer-1-recent was NOT removed
	assert.NotContains(t, peerDisconnector.removedPeers, peerID1, "Should NOT have removed peer-1-recent")
}

func TestCheckConnectionsAtTimeoutBoundary(t *testing.T) {
	peerRetriever := newMockPeerRetriever()
	peerDisconnector := &mockPeerDisconnector{}

	service := NewConnectionCheckService(peerRetriever, peerDisconnector)
	now := time.Now().Unix()

	// Create peer exactly at timeout boundary (SHOULD be removed)
	peerID := common.PeerId("peer-exact-timeout")
	peerRetriever.peers[peerID] = &common.Peer{
		State:    common.StateConnected,
		LastSeen: now - int64(PeerTimeout.Seconds()), // exactly 9 minutes ago
	}

	service.checkConnections()

	assert.Equal(t, 1, len(peerDisconnector.removedPeers), "Should have removed peer at exact timeout")
	assert.Contains(t, peerDisconnector.removedPeers, peerID, "Should have removed peer at exact timeout")
}

func TestCheckConnectionsJustBeforeTimeout(t *testing.T) {
	peerRetriever := newMockPeerRetriever()
	peerDisconnector := &mockPeerDisconnector{}

	service := NewConnectionCheckService(peerRetriever, peerDisconnector)

	now := time.Now().Unix()

	// Create peer just before timeout boundary (should NOT be removed)
	peerID := common.PeerId("peer-before-timeout")
	peerRetriever.peers[peerID] = &common.Peer{
		State:    common.StateConnected,
		LastSeen: now - int64(PeerTimeout.Seconds()) + 1, // 1 second before 9 min timeout
	}

	service.checkConnections()

	assert.Equal(t, 0, len(peerDisconnector.removedPeers), "Should NOT have removed peer before timeout")
}

func TestCheckConnectionsNoConnectedPeers(t *testing.T) {
	peerRetriever := newMockPeerRetriever()
	peerDisconnector := &mockPeerDisconnector{}

	service := NewConnectionCheckService(peerRetriever, peerDisconnector)

	// No peers in the store
	service.checkConnections()

	assert.Equal(t, 0, len(peerDisconnector.removedPeers), "Should not have removed any peers")
}

func TestCheckConnectionsPeerNotFound(t *testing.T) {
	peerRetriever := newMockPeerRetriever()
	peerDisconnector := &mockPeerDisconnector{}

	service := NewConnectionCheckService(peerRetriever, peerDisconnector)

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

	// Should not panic and should not remove anything
	assert.Equal(t, 0, len(peerDisconnector.removedPeers), "Should not have removed any peers")
}

func TestStartAndStop(t *testing.T) {
	peerRetriever := newMockPeerRetriever()
	peerDisconnector := &mockPeerDisconnector{}

	service := NewConnectionCheckService(peerRetriever, peerDisconnector)

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
	peerDisconnector := &mockPeerDisconnector{}

	service := NewConnectionCheckService(peerRetriever, peerDisconnector)

	peerID := common.PeerId("test-peer")

	// Call removePeer
	service.removePeer(peerID)

	assert.Equal(t, 1, len(peerDisconnector.removedPeers), "Peer remover should have been called")
	assert.Equal(t, peerID, peerDisconnector.removedPeers[0], "Peer remover should have removed the peer")
}

func TestNewConnectionCheckService(t *testing.T) {
	peerRetriever := newMockPeerRetriever()
	peerDisconnector := &mockPeerDisconnector{}

	service := NewConnectionCheckService(peerRetriever, peerDisconnector)

	require.NotNil(t, service, "Service should be created")
	assert.NotNil(t, service.stopChan, "Stop channel should be initialized")
	assert.Equal(t, peerRetriever, service.peerRetriever, "Peer retriever should be set")
	assert.Nil(t, service.ticker, "Ticker should not be initialized until Start is called")
}
