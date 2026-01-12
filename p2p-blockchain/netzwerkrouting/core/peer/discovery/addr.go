package discovery

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"time"

	"bjoernblessin.de/go-utils/util/logger"
)

// AddrMsgHandler defines an interface for handling incoming addr messages.
type AddrMsgHandler interface {
	HandleAddr(peerID common.PeerId, addrs []PeerAddress)
}

func (s *DiscoveryService) HandleAddr(peerID common.PeerId, addrs []PeerAddress) {
	logger.Infof("Received addr message from peer %s with %d addresses", peerID, len(addrs))

	for _, addr := range addrs {
		// Check if we already know this peer
		_, exists := s.peerRetriever.GetPeer(addr.PeerId)
		if exists {
			logger.Debugf("Already know peer %s, skipping", addr.PeerId)
			continue
		}

		// Create a new outbound peer entry for the discovered peer
		// This tracks the peer for potential future connection attempts
		newPeerID := s.peerCreator.NewOutboundPeer()
		logger.Infof("Discovered new peer: PeerId=%s (registered as %s), LastSeen=%v",
			addr.PeerId, newPeerID, time.Unix(addr.LastActiveTimestamp, 0))
	}
}
