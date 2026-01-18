package keepalive

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

//
// Mocks
//

// mockHeartbeatMsgSender is a mock implementation of HeartbeatMsgSender for testing.
type mockHeartbeatMsgSender struct {
	pingCalls []common.PeerId
	pongCalls []common.PeerId
}

func (m *mockHeartbeatMsgSender) SendHeartbeatBing(peerID common.PeerId) {
	m.pingCalls = append(m.pingCalls, peerID)
}

func (m *mockHeartbeatMsgSender) SendHeartbeatBong(peerID common.PeerId) {
	m.pongCalls = append(m.pongCalls, peerID)
}

// mockPeerRetriever is a mock implementation of PeerRetriever for testing.
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

func (m *mockPeerRetriever) GetAllOutboundPeers() []common.PeerId { // TODO
	ids := make([]common.PeerId, 0)
	for id, p := range m.peers {
		if p.State == common.StateConnected {
			ids = append(ids, id)
		}
	}
	return ids
}

//
// Tests
//

func TestHandleHeartbeatBing(t *testing.T) {
	peerRetriever := newMockPeerRetriever()
	mockSender := &mockHeartbeatMsgSender{}

	service := NewKeepaliveService(peerRetriever, mockSender)

	// Create a peer
	peerID := common.PeerId("test-peer")
	testPeer := &common.Peer{State: common.StateConnected}
	peerRetriever.peers[peerID] = testPeer

	// Initially LastSeen should be 0
	assert.Equal(t, int64(0), testPeer.LastSeen)

	// Handle heartbeat bing
	service.HandleHeartbeatBing(peerID)

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

func TestHandleHeartbeatBong(t *testing.T) {
	peerRetriever := newMockPeerRetriever()
	mockSender := &mockHeartbeatMsgSender{}

	service := NewKeepaliveService(peerRetriever, mockSender)

	// Create a peer
	peerID := common.PeerId("test-peer")
	testPeer := &common.Peer{State: common.StateConnected}
	peerRetriever.peers[peerID] = testPeer

	// Initially LastSeen should be 0
	assert.Equal(t, int64(0), testPeer.LastSeen)

	// Handle heartbeat bong
	service.HandleHeartbeatBong(peerID)

	// LastSeen should be updated
	assert.NotZero(t, testPeer.LastSeen)

	// LastSeen should be recent (within last second)
	now := time.Now().Unix()
	assert.True(t, now-testPeer.LastSeen <= 1, "LastSeen should be recent")

	// No pong should be sent (pong doesn't trigger another pong)
	assert.Equal(t, 0, len(mockSender.pongCalls))
}

func TestHandleHeartbeatBingUnknownPeer(t *testing.T) {
	peerRetriever := newMockPeerRetriever()
	mockSender := &mockHeartbeatMsgSender{}

	service := NewKeepaliveService(peerRetriever, mockSender)

	// Handle heartbeat bing for unknown peer - should not panic
	unknownPeerID := common.PeerId("unknown")
	service.HandleHeartbeatBing(unknownPeerID)

	// No pong should be sent for unknown peer
	assert.Equal(t, 0, len(mockSender.pongCalls))
}
