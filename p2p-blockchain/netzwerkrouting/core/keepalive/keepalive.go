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
	SendHeartbeatBing(peerID common.PeerId)
	SendHeartbeatBong(peerID common.PeerId)
}

// HeartbeatMsgHandler defines an interface for handling incoming heartbeat messages.
type HeartbeatMsgHandler interface {
	HandleHeartbeatBing(peerID common.PeerId)
	HandleHeartbeatBong(peerID common.PeerId)
}

// peerRetriever is an interface for retrieving peers.
// It is implemented by peer.PeerStore.
type peerRetriever interface {
	GetPeer(id common.PeerId) (*peer.Peer, bool)
	GetAllOutboundPeers() []common.PeerId
}

// KeepaliveService handles keepalive (heartbeat) functionality for peers.
// It maintains peer liveness through periodic ping/pong messages.
type KeepaliveService struct {
	peerRetriever     peerRetriever
	heartbeatSender   HeartbeatMsgSender
	stopChan          chan struct{}
	ticker            *time.Ticker
	heartbeatInterval time.Duration
}

// NewKeepaliveService creates a new KeepaliveService with a 5-minute interval.
func NewKeepaliveService(
	peerRetriever peerRetriever,
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

// sendHeartbeats sends heartbeat ping messages to all connected outbound peers.
func (s *KeepaliveService) sendHeartbeats() {
	connectedPeers := s.peerRetriever.GetAllOutboundPeers()
	logger.Debugf("Sending heartbeat bings to %d connected outbound peers", len(connectedPeers))

	for _, peerID := range connectedPeers {
		go s.heartbeatSender.SendHeartbeatBing(peerID)
	}
}

// HandleHeartbeatBing handles an incoming HeartbeatBing message from a peer.
// It updates LastSeen timestamp for the peer and responds with HeartbeatBong.
func (s *KeepaliveService) HandleHeartbeatBing(peerID common.PeerId) {
	peer, exists := s.peerRetriever.GetPeer(peerID)
	if !exists {
		logger.Warnf("Received HeartbeatBing from unknown peer %s", peerID)
		return
	}

	peer.Lock()
	now := time.Now().Unix()
	peer.LastSeen = now
	peer.Unlock()

	logger.Debugf("Received HeartbeatBing from peer %s, updated LastSeen to %v", peerID, time.Unix(now, 0))

	// Send HeartbeatBong back to the peer
	go s.heartbeatSender.SendHeartbeatBong(peerID)
}

// HandleHeartbeatBong handles an incoming HeartbeatBong message from a peer.
// It updates LastSeen timestamp for the peer that sent the original bing.
func (s *KeepaliveService) HandleHeartbeatBong(peerID common.PeerId) {
	peer, exists := s.peerRetriever.GetPeer(peerID)
	if !exists {
		logger.Warnf("Received HeartbeatBong from unknown peer %s", peerID)
		return
	}

	peer.Lock()
	now := time.Now().Unix()
	peer.LastSeen = now
	peer.Unlock()

	logger.Debugf("Received HeartbeatBong from peer %s, updated LastSeen to %v", peerID, time.Unix(now, 0))
}
