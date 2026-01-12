package discovery

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"time"

	"bjoernblessin.de/go-utils/util/assert"
	"bjoernblessin.de/go-utils/util/logger"
)

// AddrMsgHandler defines an interface for handling incoming addr messages.
type AddrMsgHandler interface {
	HandleAddr(peerID common.PeerId, addrs []PeerAddress)
}

func (s *DiscoveryService) HandleAddr(peerID common.PeerId, addrs []PeerAddress) {
	logger.Infof("Received addr message from peer %s with %d addresses", peerID, len(addrs))

	// There not much to do here because the infrastructure layer has already handled the registration of PeerIds from the received addresses.
	// With other words, the peers are already known to the PeerStore.

	for _, addr := range addrs {
		peer, exists := s.peerRetriever.GetPeer(addr.PeerId)
		assert.Assert(exists, "peer should already be registered by infrastructure layer")

		// We only need to set the last seen timestamp for the peers

		// If the last seen timestamp is older than the received one, update it
		if peer.LastSeen < addr.LastActiveTimestamp {
			peer.Lock()
			peer.LastSeen = addr.LastActiveTimestamp
			peer.Unlock()
		}

		logger.Infof("Discovered peer from Addr msg: PeerId=%s, LastSeen=%v",
			addr.PeerId, time.Unix(addr.LastActiveTimestamp, 0))
	}
}
