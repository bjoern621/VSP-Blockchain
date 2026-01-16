package keepalive

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/data/peer"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// mockHeartbeatMsgSender is a mock implementation of HeartbeatMsgSender for testing.
type mockHeartbeatMsgSender struct {
	pingCalls []common.PeerId
	pongCalls []common.PeerId
}

func (m *mockHeartbeatMsgSender) SendHeartbeatPing(peerID common.PeerId) {
	m.pingCalls = append(m.pingCalls, peerID)
}

func (m *mockHeartbeatMsgSender) SendHeartbeatPong(peerID common.PeerId) {
	m.pongCalls = append(m.pongCalls, peerID)
}

// mockPeerRetriever is a mock implementation of PeerRetriever for testing.
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

func (m *mockPeerRetriever) GetConnectedPeers() []common.PeerId {
	ids := make([]common.PeerId, 0)
	for id, p := range m.peers {
		if p.State == common.StateConnected {
			ids = append(ids, id)
		}
	}
	return ids
}

func TestHandleHeartbeatPing(t *testing.T) {
	peerRetriever := newMockPeerRetriever()
	mockSender := &mockHeartbeatMsgSender{}

	service := NewKeepaliveService(peerRetriever, mockSender)

	// Create a peer
	peerID := common.PeerId("test-peer")
	testPeer := &peer.Peer{State: common.StateConnected}
	peerRetriever.peers[peerID] = testPeer

	// Initially LastSeen should be 0
	assert.Equal(t, int64(0), testPeer.LastSeen)

	// Handle heartbeat ping
	service.HandleHeartbeatPing(peerID)

	// LastSeen should be updated
	assert.NotZero(t, testPeer.LastSeen)

	// LastSeen should be recent (within last second)
	now := time.Now().Unix()
	assert.True(t, now-testPeer.LastSeen <= 1, "LastSeen should be recent")

	// Wait for async pong to be sent
	time.Sleep(10 * time.Millisecond)

	// A pong should have been sent back
	assert.Equal(t, 1, len(mockSender.pongCalls))
	assert.Equal(t, peerID, mockSender.pongCalls[0])
}

func TestHandleHeartbeatPong(t *testing.T) {
	peerRetriever := newMockPeerRetriever()
	mockSender := &mockHeartbeatMsgSender{}

	service := NewKeepaliveService(peerRetriever, mockSender)

	// Create a peer
	peerID := common.PeerId("test-peer")
	testPeer := &peer.Peer{State: common.StateConnected}
	peerRetriever.peers[peerID] = testPeer

	// Initially LastSeen should be 0
	assert.Equal(t, int64(0), testPeer.LastSeen)

	// Handle heartbeat pong
	service.HandleHeartbeatPong(peerID)

	// LastSeen should be updated
	assert.NotZero(t, testPeer.LastSeen)

	// LastSeen should be recent (within last second)
	now := time.Now().Unix()
	assert.True(t, now-testPeer.LastSeen <= 1, "LastSeen should be recent")

	// No pong should be sent (pong doesn't trigger another pong)
	assert.Equal(t, 0, len(mockSender.pongCalls))
}

func TestHandleHeartbeatPingUnknownPeer(t *testing.T) {
	peerRetriever := newMockPeerRetriever()
	mockSender := &mockHeartbeatMsgSender{}

	service := NewKeepaliveService(peerRetriever, mockSender)

	// Handle heartbeat ping for unknown peer - should not panic
	unknownPeerID := common.PeerId("unknown")
	service.HandleHeartbeatPing(unknownPeerID)

	// No pong should be sent for unknown peer
	assert.Equal(t, 0, len(mockSender.pongCalls))
}

func TestKeepaliveServiceSendsToConnectedPeers(t *testing.T) {
	peerRetriever := newMockPeerRetriever()
	mockSender := &mockHeartbeatMsgSender{}

	// Create connected peers
	peerID1 := common.PeerId("peer-1")
	peerID2 := common.PeerId("peer-2")
	peerID3 := common.PeerId("peer-3")

	peerRetriever.peers[peerID1] = &peer.Peer{State: common.StateConnected}
	peerRetriever.peers[peerID2] = &peer.Peer{State: common.StateConnected}
	peerRetriever.peers[peerID3] = &peer.Peer{State: common.StateNew} // Not connected

	service := NewKeepaliveService(peerRetriever, mockSender)

	// Start with a very short interval for testing
	service.heartbeatInterval = 100 * time.Millisecond
	service.Start()

	// Wait for at least one tick
	time.Sleep(150 * time.Millisecond)

	// Stop the service
	service.Stop()

	// Verify heartbeat pings were sent only to connected peers
	assert.Equal(t, 2, len(mockSender.pingCalls), "should send to 2 connected peers")
	assert.Contains(t, mockSender.pingCalls, peerID1, "should send to peerID1")
	assert.Contains(t, mockSender.pingCalls, peerID2, "should send to peerID2")
	assert.NotContains(t, mockSender.pingCalls, peerID3, "should not send to peerID3")
}
