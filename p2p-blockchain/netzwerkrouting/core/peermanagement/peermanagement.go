// Package peermanagement provides automatic peer connection management.
// It monitors the connected peer count and automatically establishes new connections
// when the count falls below a configured threshold.
// Works in conjunction with the connectioncheck and keepalive packages.
//
// General Operation Flow:
//  1. Periodic Check: Every `checkInterval`, the service checks the current peer count
//  2. Threshold Evaluation: If count < `minPeers`, new connections are needed
//  3. Peer Discovery: Queries the registry for available peer IDs
//  4. Deduplication: Removes duplicate peer IDs from the list
//  5. Connection Initiation: Attempts to establish connections up to `maxPeersPerAttempt`
//  6. Handshake: For each peer, initiates the handshake process via HandshakeService
//  7. Result Logging: Logs successful/failed connection attempts
package peermanagement

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"time"

	"bjoernblessin.de/go-utils/util/logger"
)

const (
	// DefaultPeerCheckInterval is how often to check the peer count.
	DefaultPeerCheckInterval = 2 * time.Minute
	// DefaultMinPeers is the minimum number of peers to maintain.
	DefaultMinPeers = 8
	// DefaultMaxPeersPerAttempt is the maximum number of new connections to establish in a single check.
	DefaultMaxPeersPerAttempt = 3
)

// peerCounter is an interface for getting the current peer count.
type peerCounter interface {
	// GetAllConnectedPeers retrieves all connected peers' IDs (both inbound and outbound).
	GetAllConnectedPeers() []common.PeerId
}

// peerDiscoverer is an interface for discovering new peers.
type peerDiscoverer interface {
	// GetPeers queries the registry for available peer IDs.
	GetPeers(hostname string) ([]common.PeerId, error)
}

// peerCreator is an interface for creating new peers.
type peerCreator interface {
	// NewOutboundPeer creates a new outbound peer and returns its ID.
	NewOutboundPeer() common.PeerId
}

// handshakeInitiator is an interface for initiating connections to peers.
type handshakeInitiator interface {
	// InitiateHandshake initiates a handshake with a peer by its ID.
	InitiateHandshake(peerID common.PeerId) error
}

// PeerManagementService manages automatic peer connections.
type PeerManagementService struct {
	peerCounter        peerCounter
	peerDiscoverer     peerDiscoverer
	peerCreator        peerCreator
	handshakeInitiator handshakeInitiator

	minPeers           int
	maxPeersPerAttempt int
	checkInterval      time.Duration

	stopChan chan struct{}
	ticker   *time.Ticker
}

// NewPeerManagementService creates a new PeerManagementService.
func NewPeerManagementService(
	peerCounter peerCounter,
	peerDiscoverer peerDiscoverer,
	peerCreator peerCreator,
	handshakeInitiator handshakeInitiator,
) *PeerManagementService {
	return &PeerManagementService{
		peerCounter:        peerCounter,
		peerDiscoverer:     peerDiscoverer,
		peerCreator:        peerCreator,
		handshakeInitiator: handshakeInitiator,
		minPeers:           DefaultMinPeers,
		maxPeersPerAttempt: DefaultMaxPeersPerAttempt,
		checkInterval:      DefaultPeerCheckInterval,
		stopChan:           make(chan struct{}),
	}
}

// Start begins the periodic peer count monitoring.
func (s *PeerManagementService) Start() {
	s.ticker = time.NewTicker(s.checkInterval)
	go s.run()
}

// Stop halts the periodic peer count monitoring.
func (s *PeerManagementService) Stop() {
	if s.ticker != nil {
		s.ticker.Stop()
		select {
		case <-s.stopChan:
			// Channel already closed, do nothing
		default:
			// Channel not yet closed, close it
			close(s.stopChan)
		}
	}
}

// run periodically checks peer count and maintains connections.
func (s *PeerManagementService) run() {
	for {
		select {
		case <-s.ticker.C:
			s.checkAndMaintainPeers()
		case <-s.stopChan:
			return
		}
	}
}

// checkAndMaintainPeers checks if we need more peers and establishes connections if needed.
func (s *PeerManagementService) checkAndMaintainPeers() {
	currentPeers := s.peerCounter.GetAllConnectedPeers()
	peerCount := len(currentPeers)

	if peerCount >= s.minPeers {
		logger.Infof("Peer count check: %d connected peers (sufficient)", peerCount)
		return
	}

	peersNeeded := s.minPeers - peerCount
	logger.Infof("Peer count check: %d connected peers, need %d more", peerCount, peersNeeded)

	// Limit the number of new connections per attempt
	if peersNeeded > s.maxPeersPerAttempt {
		peersNeeded = s.maxPeersPerAttempt
	}

	s.establishNewPeers(peersNeeded)
}

// establishNewPeers attempts to establish connections to the specified number of peers.
func (s *PeerManagementService) establishNewPeers(count int) {
	// Query registry for available peers
	registryPeers, err := s.peerDiscoverer.GetPeers("")
	if err != nil {
		logger.Errorf("Failed to query registry for peers: %v", err)
		return
	}

	if len(registryPeers) == 0 {
		logger.Warnf("No peers available in registry")
		return
	}

	// Deduplicate peers (in case the registry returns duplicates)
	uniquePeers := s.deduplicatePeers(registryPeers)

	// Limit to the number of peers we need
	if len(uniquePeers) > count {
		uniquePeers = uniquePeers[:count]
	}

	// Attempt to establish connections
	successfulConnections := 0
	for _, peerID := range uniquePeers {
		err := s.handshakeInitiator.InitiateHandshake(peerID)
		if err != nil {
			logger.Warnf("Failed to initiate handshake with peer %s: %v", peerID, err)
			continue
		}

		successfulConnections++
		logger.Infof("Successfully initiated handshake with peer %s", peerID)
	}

	logger.Infof("Established %d/%d new peer connections", successfulConnections, count)
}

// deduplicatePeers removes duplicate peer IDs from the list.
func (s *PeerManagementService) deduplicatePeers(peers []common.PeerId) []common.PeerId {
	seen := make(map[common.PeerId]struct{})
	unique := make([]common.PeerId, 0, len(peers))

	for _, peerID := range peers {
		if _, exists := seen[peerID]; !exists {
			seen[peerID] = struct{}{}
			unique = append(unique, peerID)
		}
	}

	return unique
}
