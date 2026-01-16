// Package connectioncheck provides periodic verification of peer connections.
// It checks the LastSeen timestamp of connected peers every 16 minutes and
// removes unreachable peers to maintain a healthy P2P network.
package connectioncheck

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/data/peer"
	"time"

	"bjoernblessin.de/go-utils/util/logger"
)

const (
	// ConnectionCheckInterval is the interval at which we check peer connections (16 minutes).
	ConnectionCheckInterval = 16 * time.Minute
	// PeerTimeout is the duration after which a peer is considered unreachable if no heartbeat received.
	// Using the same interval as the check (16 minutes) means a peer that hasn't responded to
	// at least 3 heartbeat messages (5-minute intervals) will be removed.
	PeerTimeout = ConnectionCheckInterval
)

// peerRetriever is an interface for retrieving peer information.
type peerRetriever interface {
	// GetAllConnectedPeers retrieves all connected peers' IDs (both inbound and outbound).
	GetAllConnectedPeers() []common.PeerId
	// GetPeer retrieves a peer by its ID.
	GetPeer(id common.PeerId) (*peer.Peer, bool)
}

// peerRemover is an interface for removing peers.
type peerRemover interface {
	// RemovePeer removes a peer from the internal state.
	RemovePeer(id common.PeerId)
}

// ConnectionCheckService provides periodic verification of peer connections.
type ConnectionCheckService struct {
	peerRetriever      peerRetriever
	storePeerRemover   peerRemover
	networkInfoCleaner peerRemover
	stopChan           chan struct{}
	ticker             *time.Ticker
}

// NewConnectionCheckService creates a new ConnectionCheckService.
func NewConnectionCheckService(
	peerRetriever peerRetriever,
	peerRemover peerRemover,
	networkInfoCleaner peerRemover,
) *ConnectionCheckService {
	return &ConnectionCheckService{
		peerRetriever:      peerRetriever,
		storePeerRemover:   peerRemover,
		networkInfoCleaner: networkInfoCleaner,
		stopChan:           make(chan struct{}),
	}
}

// Start begins the periodic connection check loop.
func (s *ConnectionCheckService) Start() {
	logger.Infof("Starting connection check service with interval: %v", ConnectionCheckInterval)

	s.ticker = time.NewTicker(ConnectionCheckInterval)

	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.checkConnections()
			case <-s.stopChan:
				return
			}
		}
	}()
}

// Stop stops the periodic connection check loop.
func (s *ConnectionCheckService) Stop() {
	logger.Infof("Stopping connection check service")
	if s.ticker != nil {
		s.ticker.Stop()
		s.ticker = nil
	}
	select {
	case <-s.stopChan:
		// Channel already closed, do nothing
	default:
		close(s.stopChan)
	}
}

// checkConnections checks all connected peers and removes those that haven't been seen recently.
func (s *ConnectionCheckService) checkConnections() {
	logger.Debugf("Running periodic connection check")

	connectedPeers := s.peerRetriever.GetAllConnectedPeers()
	logger.Debugf("Checking %d connected peers", len(connectedPeers))

	now := time.Now()
	removedCount := 0

	for _, peerID := range connectedPeers {
		p, exists := s.peerRetriever.GetPeer(peerID)
		if !exists {
			logger.Warnf("Peer %s not found in store during connection check", peerID)
			continue
		}

		p.Lock()
		lastSeen := p.LastSeen
		peerState := p.State
		p.Unlock()

		// Check if peer is considered unreachable
		if lastSeen == 0 {
			logger.Warnf("Peer %s (state: %s) has LastSeen=0, removing", peerID, peerState)
			s.removePeer(peerID)
			removedCount++
			continue
		}

		lastSeenTime := time.Unix(lastSeen, 0)
		timeSinceLastSeen := now.Sub(lastSeenTime)

		if timeSinceLastSeen >= PeerTimeout {
			logger.Infof("Removing unreachable peer %s: last seen %v ago (threshold: %v)",
				peerID, timeSinceLastSeen.Round(time.Second), PeerTimeout)
			s.removePeer(peerID)
			removedCount++
		} else {
			logger.Tracef("Peer %s is healthy: last seen %v ago", peerID, timeSinceLastSeen.Round(time.Second))
		}
	}

	if removedCount > 0 {
		logger.Infof("Connection check completed: removed %d unreachable peers", removedCount)
	} else {
		logger.Debugf("Connection check completed: all peers healthy")
	}
}

// removePeer removes a peer from both the peer store and the network info registry.
func (s *ConnectionCheckService) removePeer(peerID common.PeerId) {
	// Remove from network info registry first (closes gRPC connection)
	s.networkInfoCleaner.RemovePeer(peerID)

	// Then remove from peer store
	s.storePeerRemover.RemovePeer(peerID)
}
