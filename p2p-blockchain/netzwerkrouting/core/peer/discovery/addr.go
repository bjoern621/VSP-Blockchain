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

	// There is not much to do here because the infrastructure layer has already handled the registration of PeerIds from the received addresses.
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

	forwardAddrs(addrs, peerID)
}

// forwardAddrs forwards the given addrs to neighboring peers.
// Forwarding rules:
//   - Do not forward to the peer from which we received the addr.
//   - Do not forward peers that were already known (how do we detect that? i think we need to change getorregisterpeer or so, so that peer store can distinguish between "created by infrastructure" and "really created"? but how should be do that? we want to keep a strong seperation between infrastructure and core/domain...).
//   - ... more
func forwardAddrs(addrs []PeerAddress, sender common.PeerId) {
}

// selectPeersForAddrForwarding selects up to maxAddrs unique peers from the provided list for addr message forwarding.
func selectPeersForAddrForwarding(peers []common.PeerId, maxAddrs int) []common.PeerId {
	if len(peers) <= maxAddrs {
		return peers
	}

	selected := make([]common.PeerId, 0, maxAddrs)
	peerSet := make(map[common.PeerId]struct{})

	for len(selected) < maxAddrs {
		for _, peerID := range peers {
			if len(selected) >= maxAddrs {
				break
			}
			if _, exists := peerSet[peerID]; !exists {
				peerSet[peerID] = struct{}{}
				selected = append(selected, peerID)
			}
		}
	}

	return selected
}
