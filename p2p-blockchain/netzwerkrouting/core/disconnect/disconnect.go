package disconnect

import (
	"fmt"
	"time"

	"s3b/vsp-blockchain/p2p-blockchain/internal/common"

	"bjoernblessin.de/go-utils/util/logger"
)

// HolddownDuration is the time a peer remains in holddown state before being fully removed.
// During holddown, the peer rejects new connection attempts.
const HolddownDuration = 15 * time.Minute

// DisconnectService defines the interface for disconnecting/forgetting peers.
type DisconnectService interface {
	// Disconnect puts a peer into holddown state.
	//
	// The peer's gRPC connection is closed immediately, but the peer remains in the store
	// with StateHolddown for the holddown duration (15 minutes). During holddown:
	//   - New connection attempts from this peer are rejected
	//   - The peer is excluded from peer lists (connected peers, discovery, etc.)
	//   - After holddown expires, the peer is fully removed by ConnectionCheckService
	//
	// Returns an error if the peer does not exist.
	Disconnect(peerID common.PeerId) error
}

// connectionCloser is implemented by infrastructure layer (NetworkInfoRegistry) to close
// connections while preserving address mappings for holddown detection.
type connectionCloser interface {
	// CloseConnection closes the gRPC connection but preserves address mappings.
	CloseConnection(id common.PeerId)
}

// peerRetriever is an interface for retrieving peers.
// It is implemented by peer.PeerStore.
type peerRetriever interface {
	GetPeer(id common.PeerId) (*common.Peer, bool)
}

// disconnectService implements DisconnectService with the actual domain logic.
type disconnectService struct {
	connectionCloser connectionCloser
	peerRetriever    peerRetriever
}

func NewDisconnectService(
	connectionCloser connectionCloser,
	peerStore peerRetriever,
) DisconnectService {
	return &disconnectService{
		connectionCloser: connectionCloser,
		peerRetriever:    peerStore,
	}
}

// Disconnect puts a peer into holddown state.
//
// This involves:
//  1. Verifying the peer exists in the store
//  2. Closing the gRPC connection (but keeping address mappings for rejection detection)
//  3. Setting peer state to StateHolddown with HolddownStartTime
//
// The peer will remain in holddown for 15 minutes, during which:
//   - All incoming connection attempts are rejected with REJECT_HOLDDOWN
//   - The peer is excluded from GetAllOutboundPeers, GetUnconnectedPeers, etc.
//
// After 15 minutes, ConnectionCheckService will fully remove the peer.
func (s *disconnectService) Disconnect(peerID common.PeerId) error {
	peer, ok := s.peerRetriever.GetPeer(peerID)
	if !ok {
		return fmt.Errorf("peer %s not found in store", peerID)
	}

	peer.Lock()
	currentState := peer.State
	if currentState == common.StateHolddown {
		peer.Unlock()
		logger.Debugf("[disconnect] Peer %s is already in holddown state", peerID)
		return nil
	}
	peer.State = common.StateHolddown
	peer.HolddownStartTime = time.Now().Unix()
	peer.Unlock()

	logger.Infof("[disconnect] Peer %s entering holddown state (was: %s)", peerID, currentState)

	// Close gRPC connection but keep address mappings for rejection detection
	s.connectionCloser.CloseConnection(peerID)

	return nil
}
