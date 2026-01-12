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

		// Core layer only tracks peer IDs
		logger.Debugf("New peer discovered: PeerId=%s, LastSeen=%v",
			addr.PeerId, time.Unix(addr.LastActiveTimestamp, 0))
		// TODO: Trigger peer creation and connection for new peer
	}
}
