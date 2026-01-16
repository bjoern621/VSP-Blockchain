// Package peermanagement watches active peer connection count.
// It monitors the connected peer count and automatically establishes new connections
// when the count falls below a configured threshold.
// Works in conjunction with the connectioncheck and keepalive packages as well as the gossip discovery service.
//
// General Operation Flow:
//  1. Periodic Check: Every `checkInterval`, the service checks the current peer count
//  2. Threshold Evaluation: If count < `minPeers`, new connections are needed
//  3. Connection Initiation: Attempts to establish connections up to `maxPeersPerAttempt` using `GetUnconnectedPeers()`
//  4. Handshake: For each peer, initiates the handshake process via HandshakeService
//
// Note: Peer discovery is handled separately:
//   - Bootstrap: Registry query at startup (discovery.GetPeers)
//   - Gossip: Periodic getaddr to peers (discovery.GossipDiscoveryService)
//
// This service focuses only on connections to already discovered peers.
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
	GetPeers()
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

type knownPeerRetriever interface {
	// GetUnconnectedPeers retrieves peer IDs that are known but not currently connected.
	GetUnconnectedPeers() []common.PeerId
}

// PeerManagementService manages automatic peer connections.
type PeerManagementService struct {
	peerCounter        peerCounter
	peerDiscoverer     peerDiscoverer
	peerCreator        peerCreator
	handshakeInitiator handshakeInitiator
	knownPeerRetriever knownPeerRetriever

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
	knownPeerRetriever knownPeerRetriever,
) *PeerManagementService {
	return &PeerManagementService{
		peerCounter:        peerCounter,
		peerDiscoverer:     peerDiscoverer,
		peerCreator:        peerCreator,
		handshakeInitiator: handshakeInitiator,
		knownPeerRetriever: knownPeerRetriever,
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
	potentialPeers := s.knownPeerRetriever.GetUnconnectedPeers()

	if len(potentialPeers) == 0 {
		logger.Warnf("No unconnected peers available")
		return
	}

	// Limit to the number of peers we need
	if len(potentialPeers) > count {
		potentialPeers = potentialPeers[:count]
		// TODO select random
	}

	// Attempt to establish connections
	successfulConnections := 0
	for _, peerID := range potentialPeers {
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
