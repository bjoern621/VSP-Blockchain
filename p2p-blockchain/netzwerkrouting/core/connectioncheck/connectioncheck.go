// Package connectioncheck provides periodic verification of peer connections.
// It checks the LastSeen timestamp of connected peers every 16 minutes and
// removes unreachable peers to maintain a healthy P2P network.
package connectioncheck

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"time"

	"bjoernblessin.de/go-utils/util/logger"
)

const (
	// ConnectionCheckInterval is the interval at which we check peer connections.
	ConnectionCheckInterval = 10 * time.Minute
	// PeerTimeout is the duration after which a peer is considered unreachable if no heartbeat received.
	// Using 9 minutes means a peer that hasn't responded to
	// at least 2 heartbeat messages (4-minutes intervals) will be removed.
	PeerTimeout = 9 * time.Minute
)

// peerRetriever is an interface for retrieving peer information.
type peerRetriever interface {
	// GetPeersWithHandshakeStarted retrieves all peers' IDs that have started the handshake process.
	GetPeersWithHandshakeStarted() []common.PeerId
	// GetPeer retrieves a peer by its ID.
	GetPeer(id common.PeerId) (*common.Peer, bool)
}

type peerDisconnector interface {
	// DisconnectPeer disconnects a peer by its ID.
	Disconnect(id common.PeerId) error
}

// ConnectionCheckService provides periodic verification of peer connections.
type ConnectionCheckService struct {
	peerRetriever    peerRetriever
	peerDisconnector peerDisconnector
	stopChan         chan struct{}
	ticker           *time.Ticker
}

// NewConnectionCheckService creates a new ConnectionCheckService.
func NewConnectionCheckService(
	peerRetriever peerRetriever,
	peerDisconnector peerDisconnector,
) *ConnectionCheckService {
	return &ConnectionCheckService{
		peerRetriever:    peerRetriever,
		peerDisconnector: peerDisconnector,
		stopChan:         make(chan struct{}),
	}
}

// Start begins the periodic connection check loop.
func (s *ConnectionCheckService) Start() {
	logger.Infof("[conn-check] Starting connection check service with interval: %v", ConnectionCheckInterval)

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
	logger.Infof("[conn-check] Stopping connection check service")
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
	logger.Debugf("[conn-check] Running periodic connection check")

	connectedPeers := s.peerRetriever.GetPeersWithHandshakeStarted()
	logger.Debugf("[conn-check] Checking %d connected peers", len(connectedPeers))

	now := time.Now()
	removedCount := 0

	for _, peerID := range connectedPeers {
		p, exists := s.peerRetriever.GetPeer(peerID)
		if !exists {
			logger.Warnf("[conn-check] Peer %s not found in store during connection check", peerID)
			continue
		}

		p.Lock()
		lastSeen := p.LastSeen
		peerState := p.State
		p.Unlock()

		// Check if peer is considered unreachable
		if lastSeen == 0 {
			logger.Warnf("[conn-check] Peer %s (state: %s) has LastSeen=0, removing", peerID, peerState)
			s.removePeer(peerID)
			removedCount++
			continue
		}

		lastSeenTime := time.Unix(lastSeen, 0)
		timeSinceLastSeen := now.Sub(lastSeenTime)

		if timeSinceLastSeen >= PeerTimeout {
			logger.Infof("[conn-check] Removing unreachable peer %s: last seen %v ago (threshold: %v)",
				peerID, timeSinceLastSeen.Round(time.Second), PeerTimeout)
			s.removePeer(peerID)
			removedCount++
		} else {
			logger.Debugf("[conn-check] Peer %s is healthy: last seen %v ago", peerID, timeSinceLastSeen.Round(time.Second))
		}
	}

	if removedCount > 0 {
		logger.Infof("[conn-check] Connection check completed: removed %d unreachable peers", removedCount)
	} else {
		logger.Debugf("[conn-check] Connection check completed: all peers healthy")
	}
}

// removePeer removes a peer from both the peer store and the network info registry.
func (s *ConnectionCheckService) removePeer(peerID common.PeerId) {
	_ = s.peerDisconnector.Disconnect(peerID)
}
