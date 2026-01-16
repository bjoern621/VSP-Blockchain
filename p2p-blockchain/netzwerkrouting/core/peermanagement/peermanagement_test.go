package peermanagement

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

// mockPeerCounter is a mock implementation of peerCounter for testing.
type mockPeerCounter struct {
	peerIDs []common.PeerId
}

func (m *mockPeerCounter) GetAllConnectedPeers() []common.PeerId {
	return m.peerIDs
}

// mockPeerDiscoverer is a mock implementation of peerDiscoverer for testing.
type mockPeerDiscoverer struct {
	peerIDs  []common.PeerId
	queryErr error
}

func (m *mockPeerDiscoverer) GetPeers(hostname string) ([]common.PeerId, error) {
	if m.queryErr != nil {
		return nil, m.queryErr
	}
	return m.peerIDs, nil
}

// mockPeerCreator is a mock implementation of peerCreator for testing.
type mockPeerCreator struct {
	peerIDCounter int
	peerIDs       []common.PeerId
}

func (m *mockPeerCreator) NewOutboundPeer() common.PeerId {
	peerID := common.PeerId("mock-peer-" + string(rune(m.peerIDCounter)))
	m.peerIDs = append(m.peerIDs, peerID)
	m.peerIDCounter++
	return peerID
}

// mockHandshakeInitiator is a mock implementation of handshakeInitiator for testing.
type mockHandshakeInitiator struct {
	initiatedPeers []common.PeerId
	connectErrors  map[common.PeerId]bool
}

func (m *mockHandshakeInitiator) InitiateHandshake(peerID common.PeerId) error {
	if m.connectErrors == nil {
		m.connectErrors = make(map[common.PeerId]bool)
	}

	if m.connectErrors[peerID] {
		return assert.AnError
	}

	m.initiatedPeers = append(m.initiatedPeers, peerID)
	return nil
}

//
// Tests
//

func TestNewPeerManagementService(t *testing.T) {
	peerCounter := &mockPeerCounter{}
	peerDiscoverer := &mockPeerDiscoverer{}
	peerCreator := &mockPeerCreator{}
	handshakeInitiator := &mockHandshakeInitiator{}

	service := NewPeerManagementService(peerCounter, peerDiscoverer, peerCreator, handshakeInitiator)

	require.NotNil(t, service, "Service should be created")
	assert.NotNil(t, service.stopChan, "Stop channel should be initialized")
	assert.Equal(t, DefaultMinPeers, service.minPeers, "Should use default minPeers")
	assert.Equal(t, DefaultMaxPeersPerAttempt, service.maxPeersPerAttempt, "Should use default maxPeersPerAttempt")
	assert.Equal(t, DefaultPeerCheckInterval, service.checkInterval, "Should use default checkInterval")
	assert.Equal(t, peerCounter, service.peerCounter, "Peer counter should be set")
	assert.Equal(t, peerDiscoverer, service.peerDiscoverer, "Peer discoverer should be set")
	assert.Equal(t, peerCreator, service.peerCreator, "Peer creator should be set")
	assert.Equal(t, handshakeInitiator, service.handshakeInitiator, "Handshake initiator should be set")
	assert.Nil(t, service.ticker, "Ticker should not be initialized until Start is called")
}

func TestSetMinPeers(t *testing.T) {
	peerCounter := &mockPeerCounter{}
	peerDiscoverer := &mockPeerDiscoverer{}
	peerCreator := &mockPeerCreator{}
	handshakeInitiator := &mockHandshakeInitiator{}

	service := NewPeerManagementService(peerCounter, peerDiscoverer, peerCreator, handshakeInitiator)

	service.minPeers = 12
	assert.Equal(t, 12, service.minPeers, "minPeers should be updated")
}

func TestSetMaxPeersPerAttempt(t *testing.T) {
	peerCounter := &mockPeerCounter{}
	peerDiscoverer := &mockPeerDiscoverer{}
	peerCreator := &mockPeerCreator{}
	handshakeInitiator := &mockHandshakeInitiator{}

	service := NewPeerManagementService(peerCounter, peerDiscoverer, peerCreator, handshakeInitiator)

	service.maxPeersPerAttempt = 5
	assert.Equal(t, 5, service.maxPeersPerAttempt, "maxPeersPerAttempt should be updated")
}

func TestSetCheckInterval(t *testing.T) {
	peerCounter := &mockPeerCounter{}
	peerDiscoverer := &mockPeerDiscoverer{}
	peerCreator := &mockPeerCreator{}
	handshakeInitiator := &mockHandshakeInitiator{}

	service := NewPeerManagementService(peerCounter, peerDiscoverer, peerCreator, handshakeInitiator)

	service.checkInterval = 5 * time.Minute
	assert.Equal(t, 5*time.Minute, service.checkInterval, "checkInterval should be updated")
}

func TestCheckAndMaintainPeers_SufficientPeers(t *testing.T) {
	peerCounter := &mockPeerCounter{
		peerIDs: []common.PeerId{"peer-1", "peer-2", "peer-3", "peer-4", "peer-5", "peer-6", "peer-7", "peer-8"},
	}
	peerDiscoverer := &mockPeerDiscoverer{}
	peerCreator := &mockPeerCreator{}
	handshakeInitiator := &mockHandshakeInitiator{}

	service := NewPeerManagementService(peerCounter, peerDiscoverer, peerCreator, handshakeInitiator)
	service.minPeers = 8

	service.checkAndMaintainPeers()

	// Should not attempt to connect any new peers
	assert.Equal(t, 0, len(handshakeInitiator.initiatedPeers), "Should not connect any new peers")
}

func TestCheckAndMaintainPeers_NeedsMorePeers(t *testing.T) {
	peerCounter := &mockPeerCounter{
		peerIDs: []common.PeerId{"peer-1", "peer-2", "peer-3", "peer-4", "peer-5"},
	}
	peerDiscoverer := &mockPeerDiscoverer{
		peerIDs: []common.PeerId{"registry-peer-1", "registry-peer-2", "registry-peer-3"},
	}
	peerCreator := &mockPeerCreator{}
	handshakeInitiator := &mockHandshakeInitiator{}

	service := NewPeerManagementService(peerCounter, peerDiscoverer, peerCreator, handshakeInitiator)
	service.minPeers = 8

	service.checkAndMaintainPeers()

	// Should attempt to connect 3 peers (8 - 5 = 3 needed)
	assert.Equal(t, 3, len(handshakeInitiator.initiatedPeers), "Should connect 3 new peers")
}

func TestCheckAndMaintainPeers_LimitedByMaxPerAttempt(t *testing.T) {
	peerCounter := &mockPeerCounter{
		peerIDs: []common.PeerId{"peer-1", "peer-2"},
	}
	peerDiscoverer := &mockPeerDiscoverer{
		peerIDs: []common.PeerId{"registry-peer-1", "registry-peer-2", "registry-peer-3", "registry-peer-4", "registry-peer-5"},
	}
	peerCreator := &mockPeerCreator{}
	handshakeInitiator := &mockHandshakeInitiator{}

	service := NewPeerManagementService(peerCounter, peerDiscoverer, peerCreator, handshakeInitiator)
	service.minPeers = 8
	service.maxPeersPerAttempt = 2 // Limit to 2 connections per attempt

	service.checkAndMaintainPeers()

	// Should only attempt to connect 2 peers (limited by maxPeersPerAttempt)
	assert.Equal(t, 2, len(handshakeInitiator.initiatedPeers), "Should connect only 2 peers due to limit")
}

func TestCheckAndMaintainPeers_NoPeersAvailable(t *testing.T) {
	peerCounter := &mockPeerCounter{
		peerIDs: []common.PeerId{"peer-1"},
	}
	peerDiscoverer := &mockPeerDiscoverer{
		peerIDs: []common.PeerId{},
	}
	peerCreator := &mockPeerCreator{}
	handshakeInitiator := &mockHandshakeInitiator{}

	service := NewPeerManagementService(peerCounter, peerDiscoverer, peerCreator, handshakeInitiator)
	service.minPeers = 8

	service.checkAndMaintainPeers()

	// Should not attempt to connect any peers since none are available
	assert.Equal(t, 0, len(handshakeInitiator.initiatedPeers), "Should not connect any peers when none available")
}

func TestEstablishNewPeers_SuccessfulConnections(t *testing.T) {
	peerCounter := &mockPeerCounter{}
	peerDiscoverer := &mockPeerDiscoverer{
		peerIDs: []common.PeerId{"registry-peer-1", "registry-peer-2", "registry-peer-3"},
	}
	peerCreator := &mockPeerCreator{}
	handshakeInitiator := &mockHandshakeInitiator{}

	service := NewPeerManagementService(peerCounter, peerDiscoverer, peerCreator, handshakeInitiator)

	service.establishNewPeers(3)

	assert.Equal(t, 3, len(handshakeInitiator.initiatedPeers), "Should connect 3 peers")
	assert.Equal(t, "registry-peer-1", string(handshakeInitiator.initiatedPeers[0]), "First connection should be to first peer")
	assert.Equal(t, "registry-peer-2", string(handshakeInitiator.initiatedPeers[1]), "Second connection should be to second peer")
	assert.Equal(t, "registry-peer-3", string(handshakeInitiator.initiatedPeers[2]), "Third connection should be to third peer")
}

func TestEstablishNewPeers_WithConnectionErrors(t *testing.T) {
	peerCounter := &mockPeerCounter{}
	peerDiscoverer := &mockPeerDiscoverer{
		peerIDs: []common.PeerId{"registry-peer-1", "registry-peer-2", "registry-peer-3"},
	}
	peerCreator := &mockPeerCreator{}
	handshakeInitiator := &mockHandshakeInitiator{
		connectErrors: map[common.PeerId]bool{
			"registry-peer-2": true, // Second peer fails
		},
	}

	service := NewPeerManagementService(peerCounter, peerDiscoverer, peerCreator, handshakeInitiator)

	service.establishNewPeers(3)

	// Should attempt all 3 connections, but only 2 succeed
	assert.Equal(t, 2, len(handshakeInitiator.initiatedPeers), "Should have 2 successful connections")
	assert.Equal(t, "registry-peer-1", string(handshakeInitiator.initiatedPeers[0]), "First connection should succeed")
	assert.Equal(t, "registry-peer-3", string(handshakeInitiator.initiatedPeers[1]), "Third connection should succeed")
}

func TestDeduplicatePeers(t *testing.T) {
	peerCounter := &mockPeerCounter{}
	peerDiscoverer := &mockPeerDiscoverer{}
	peerCreator := &mockPeerCreator{}
	handshakeInitiator := &mockHandshakeInitiator{}

	service := NewPeerManagementService(peerCounter, peerDiscoverer, peerCreator, handshakeInitiator)

	peers := []common.PeerId{
		"peer-1",
		"peer-1", // Duplicate
		"peer-2",
		"peer-3",
		"peer-2", // Duplicate
	}

	unique := service.deduplicatePeers(peers)

	assert.Equal(t, 3, len(unique), "Should have 3 unique peers")
	assert.Equal(t, "peer-1", string(unique[0]))
	assert.Equal(t, "peer-2", string(unique[1]))
	assert.Equal(t, "peer-3", string(unique[2]))
}

func TestStartAndStop(t *testing.T) {
	peerCounter := &mockPeerCounter{}
	peerDiscoverer := &mockPeerDiscoverer{}
	peerCreator := &mockPeerCreator{}
	handshakeInitiator := &mockHandshakeInitiator{}

	service := NewPeerManagementService(peerCounter, peerDiscoverer, peerCreator, handshakeInitiator)

	// Start the service
	service.Start()
	assert.NotNil(t, service.ticker, "Ticker should be initialized")

	// Wait a bit to ensure goroutine starts
	time.Sleep(10 * time.Millisecond)

	// Stop the service
	service.Stop()

	// Verify that calling Stop again doesn't panic
	service.Stop()
}

func TestDeduplicatePeers_EmptyList(t *testing.T) {
	peerCounter := &mockPeerCounter{}
	peerDiscoverer := &mockPeerDiscoverer{}
	peerCreator := &mockPeerCreator{}
	handshakeInitiator := &mockHandshakeInitiator{}

	service := NewPeerManagementService(peerCounter, peerDiscoverer, peerCreator, handshakeInitiator)

	peers := []common.PeerId{}
	unique := service.deduplicatePeers(peers)

	assert.Equal(t, 0, len(unique), "Should handle empty list")
}

func TestDeduplicatePeers_AllUnique(t *testing.T) {
	peerCounter := &mockPeerCounter{}
	peerDiscoverer := &mockPeerDiscoverer{}
	peerCreator := &mockPeerCreator{}
	handshakeInitiator := &mockHandshakeInitiator{}

	service := NewPeerManagementService(peerCounter, peerDiscoverer, peerCreator, handshakeInitiator)

	peers := []common.PeerId{"peer-1", "peer-2", "peer-3"}

	unique := service.deduplicatePeers(peers)

	assert.Equal(t, 3, len(unique), "Should have 3 unique peers")
	assert.Equal(t, peers, unique, "Should return same list when all unique")
}
