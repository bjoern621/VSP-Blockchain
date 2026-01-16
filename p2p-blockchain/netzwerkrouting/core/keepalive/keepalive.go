package keepalive

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/data/peer"
	"time"

	"bjoernblessin.de/go-utils/util/logger"
)

const DefaultHeartbeatInterval = 5 * time.Minute

// HeartbeatMsgSender defines an interface for sending heartbeat messages to peers.
type HeartbeatMsgSender interface {
	SendHeartbeatPing(peerID common.PeerId)
	SendHeartbeatPong(peerID common.PeerId)
}

// HeartbeatMsgHandler defines an interface for handling incoming heartbeat messages.
type HeartbeatMsgHandler interface {
	HandleHeartbeatPing(peerID common.PeerId)
	HandleHeartbeatPong(peerID common.PeerId)
}

// PeerRetriever is an interface for retrieving peers for keepalive purposes.
type PeerRetriever interface {
	// GetPeer retrieves a peer by its ID.
	GetPeer(id common.PeerId) (*peer.Peer, bool)
	// GetConnectedPeers retrieves all peers that are in Connected state.
	GetConnectedPeers() []common.PeerId
}

// KeepaliveService handles keepalive (heartbeat) functionality for peers.
// It maintains peer liveness through periodic ping/pong messages.
type KeepaliveService struct {
	peerRetriever     PeerRetriever
	heartbeatSender   HeartbeatMsgSender
	stopChan          chan struct{}
	ticker            *time.Ticker
	heartbeatInterval time.Duration
}

// NewKeepaliveService creates a new KeepaliveService with a 5-minute interval.
func NewKeepaliveService(
	peerRetriever PeerRetriever,
	heartbeatSender HeartbeatMsgSender,
) *KeepaliveService {
	return &KeepaliveService{
		peerRetriever:     peerRetriever,
		heartbeatSender:   heartbeatSender,
		stopChan:          make(chan struct{}),
		heartbeatInterval: DefaultHeartbeatInterval,
	}
}

// Start begins the keepalive service.
// It runs in a goroutine and sends heartbeat pings to all connected peers at regular intervals.
func (s *KeepaliveService) Start() {
	logger.Infof("Starting KeepaliveService with %s interval", s.heartbeatInterval)

	s.ticker = time.NewTicker(s.heartbeatInterval)

	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.sendHeartbeats()
			case <-s.stopChan:
				s.ticker.Stop()
				logger.Infof("KeepaliveService stopped")
				return
			}
		}
	}()
}

// Stop stops the keepalive service.
func (s *KeepaliveService) Stop() {
	close(s.stopChan)
}

// sendHeartbeats sends heartbeat ping messages to all connected peers.
func (s *KeepaliveService) sendHeartbeats() {
	connectedPeers := s.peerRetriever.GetConnectedPeers()
	logger.Debugf("Sending heartbeat pings to %d connected peers", len(connectedPeers))

	for _, peerID := range connectedPeers {
		s.heartbeatSender.SendHeartbeatPing(peerID)
	}
}

// HandleHeartbeatPing handles an incoming HeartbeatPing message from a peer.
// It updates LastSeen timestamp for the peer and responds with HeartbeatPong.
func (s *KeepaliveService) HandleHeartbeatPing(peerID common.PeerId) {
	peer, exists := s.peerRetriever.GetPeer(peerID)
	if !exists {
		logger.Warnf("Received HeartbeatPing from unknown peer %s", peerID)
		return
	}

	peer.Lock()
	now := time.Now().Unix()
	peer.LastSeen = now
	peer.Unlock()

	logger.Debugf("Received HeartbeatPing from peer %s, updated LastSeen to %v", peerID, time.Unix(now, 0))

	// Send HeartbeatPong back to the peer
	go s.heartbeatSender.SendHeartbeatPong(peerID)
}

// HandleHeartbeatPong handles an incoming HeartbeatPong message from a peer.
// It updates LastSeen timestamp for the peer that sent the original ping.
func (s *KeepaliveService) HandleHeartbeatPong(peerID common.PeerId) {
	peer, exists := s.peerRetriever.GetPeer(peerID)
	if !exists {
		logger.Warnf("Received HeartbeatPong from unknown peer %s", peerID)
		return
	}

	peer.Lock()
	now := time.Now().Unix()
	peer.LastSeen = now
	peer.Unlock()

	logger.Debugf("Received HeartbeatPong from peer %s, updated LastSeen to %v", peerID, time.Unix(now, 0))
}
