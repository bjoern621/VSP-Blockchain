// Package connectioncheck provides periodic verification of peer connections.
// It checks the LastSeen timestamp of connected peers every 10 minutes and
// removes unreachable peers to maintain a healthy P2P network.
// It also handles cleanup of holddown peers after the holddown period expires.
package connectioncheck

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/disconnect"
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
	// GetHolddownPeers retrieves all peers in holddown state.
	GetHolddownPeers() []common.PeerId
}

// peerRemover is an interface for permanently removing peers.
// Used for final cleanup of holddown peers after the holddown period expires.
type peerRemover interface {
	RemovePeer(id common.PeerId)
}

// networkInfoRemover is an interface for removing peers from network info registry.
// Used for final cleanup of holddown peers.
type networkInfoRemover interface {
	RemovePeer(id common.PeerId)
}

type peerDisconnector interface {
	// Disconnect puts a peer into holddown state.
	Disconnect(id common.PeerId) error
}

// ConnectionCheckService provides periodic verification of peer connections.
// It performs two types of checks:
//  1. Connected peer health: Checks LastSeen timestamps and puts unreachable peers into holddown
//  2. Holddown cleanup: Permanently removes peers whose holddown period has expired
type ConnectionCheckService struct {
	peerRetriever      peerRetriever
	peerDisconnector   peerDisconnector
	peerRemover        peerRemover
	networkInfoRemover networkInfoRemover
	stopChan           chan struct{}
	ticker             *time.Ticker
}

// NewConnectionCheckService creates a new ConnectionCheckService.
//
// Parameters:
//   - peerRetriever: Provides peer lookup (PeerStore)
//   - peerDisconnector: Puts peers into holddown (DisconnectService)
//   - peerRemover: Permanently removes peers from store (PeerStore)
//   - networkInfoRemover: Removes peers from network registry (NetworkInfoRegistry)
func NewConnectionCheckService(
	peerRetriever peerRetrieverWithHolddown,
	peerDisconnector peerDisconnector,
	networkInfoRemover networkInfoRemover,
) *ConnectionCheckService {
	return &ConnectionCheckService{
		peerRetriever:      peerRetriever,
		peerDisconnector:   peerDisconnector,
		peerRemover:        peerRetriever,
		networkInfoRemover: networkInfoRemover,
		stopChan:           make(chan struct{}),
	}
}

// peerRetrieverWithHolddown combines peerRetriever and peerRemover interfaces.
type peerRetrieverWithHolddown interface {
	peerRetriever
	peerRemover
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
				s.cleanupHolddownPeers()
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

// cleanupHolddownPeers permanently removes peers whose holddown period has expired.
// Peers are put into holddown state by DisconnectService.Disconnect() and remain there
// for HolddownDuration (15 minutes). After expiration, they are fully removed from
// both the peer store and network info registry.
func (s *ConnectionCheckService) cleanupHolddownPeers() {
	holddownPeers := s.peerRetriever.GetHolddownPeers()
	if len(holddownPeers) == 0 {
		return
	}

	logger.Debugf("[conn-check] Checking %d holddown peers for cleanup", len(holddownPeers))

	now := time.Now()
	cleanedCount := 0

	for _, peerID := range holddownPeers {
		p, exists := s.peerRetriever.GetPeer(peerID)
		if !exists {
			continue
		}

		p.Lock()
		holddownStart := p.HolddownStartTime
		p.Unlock()

		if holddownStart == 0 {
			// No holddown start time set, clean up immediately
			logger.Warnf("[conn-check] Holddown peer %s has no HolddownStartTime, removing", peerID)
			s.permanentlyRemovePeer(peerID)
			cleanedCount++
			continue
		}

		holddownStartTime := time.Unix(holddownStart, 0)
		timeSinceHolddown := now.Sub(holddownStartTime)

		if timeSinceHolddown >= disconnect.HolddownDuration {
			logger.Infof("[conn-check] Holddown expired for peer %s: in holddown for %v (threshold: %v), permanently removing",
				peerID, timeSinceHolddown.Round(time.Second), disconnect.HolddownDuration)
			s.permanentlyRemovePeer(peerID)
			cleanedCount++
		} else {
			remaining := disconnect.HolddownDuration - timeSinceHolddown
			logger.Debugf("[conn-check] Peer %s still in holddown: %v remaining", peerID, remaining.Round(time.Second))
		}
	}

	if cleanedCount > 0 {
		logger.Infof("[conn-check] Holddown cleanup completed: permanently removed %d peers", cleanedCount)
	}
}

// removePeer puts a peer into holddown state (soft-delete).
func (s *ConnectionCheckService) removePeer(peerID common.PeerId) {
	_ = s.peerDisconnector.Disconnect(peerID)
}

// permanentlyRemovePeer removes a peer from both the peer store and network info registry.
// This is the final cleanup after holddown expires.
func (s *ConnectionCheckService) permanentlyRemovePeer(peerID common.PeerId) {
	s.networkInfoRemover.RemovePeer(peerID)
	s.peerRemover.RemovePeer(peerID)
}
