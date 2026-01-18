package discovery

import (
	"math/rand"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"slices"
	"time"

	"bjoernblessin.de/go-utils/util/assert"
	"bjoernblessin.de/go-utils/util/logger"
)

// AddrMsgHandler defines an interface for handling incoming addr messages.
type AddrMsgHandler interface {
	HandleAddr(peerID common.PeerId, addrs []PeerAddress)
}

func (s *DiscoveryService) HandleAddr(peerID common.PeerId, addrs []PeerAddress) {
	peer, exists := s.peerRetriever.GetPeer(peerID)
	if !exists {
		logger.Warnf("[addr_handler] Received addr from unknown peer %s", peerID)
		return
	}

	if peer.State != common.StateConnected {
		logger.Warnf("[addr_handler] Received addr from peer %s which is not connected (state: %v)", peerID, peer.State)
		return
	}

	logger.Tracef("[addr_handler] Received addr message from peer %s with %d addresses", peerID, len(addrs))

	// There is not much to do here because the infrastructure layer has already handled the registration of PeerIds from the received addresses.
	// With other words, the peers are already known to the PeerStore.

	// Update last seen timestamps for received addresses
	for _, addr := range addrs {
		peer, exists := s.peerRetriever.GetPeer(addr.PeerId)
		assert.Assert(exists, "peer should already be registered by infrastructure layer")

		if peer.State != common.StateNew {
			continue // Only update LastSeen for peers in StateNew via discovery
		}

		if peer.LastSeen < addr.LastActiveTimestamp {
			peer.Lock()
			peer.LastSeen = addr.LastActiveTimestamp
			peer.Unlock()
		}

		logger.Tracef("[addr_handler] Discovered peer from Addr msg: PeerId=%s, LastSeen=%v",
			addr.PeerId, time.Unix(addr.LastActiveTimestamp, 0))
	}

	// Forward addresses to random neighbors
	// Uses addrs instead of filteredAddrs by design
	s.forwardAddrs(addrs, peerID)
}

// forwardAddrs forwards the given addrs to neighboring peers.
// Forwarding rules:
//   - Do not forward to the peer from which we received the addr
//   - Do not forward an address to a peer that has already received it
//   - For each address, independently select 2 random peers from connected peers to forward to
func (s *DiscoveryService) forwardAddrs(addrs []PeerAddress, sender common.PeerId) {
	connectedPeers := s.peerRetriever.GetAllOutboundPeers()

	// Filter out the sender
	eligiblePeers := slices.DeleteFunc(connectedPeers, func(peerID common.PeerId) bool {
		return peerID == sender
	})

	if len(eligiblePeers) == 0 {
		return
	}

	// For each address, independently select 2 random peers and forward
	for _, addr := range addrs {
		numPeers := min(len(eligiblePeers), 2)
		selectedPeers := selectPeersForAddrForwarding(eligiblePeers, numPeers)

		// Forward this address to selected peers if they haven't received it yet
		for _, recipientID := range selectedPeers {
			recipient, exists := s.peerRetriever.GetPeer(recipientID)
			if !exists {
				continue
			}

			if recipientID == addr.PeerId {
				// Don't forward the address back to the peer itself
				continue
			}

			recipient.Lock()

			// Check if this recipient has already received this address
			if !recipient.AddrsSentTo.Contains(addr.PeerId) {
				recipient.AddrsSentTo.Add(addr.PeerId)
				recipient.Unlock()

				go s.addrMsgSender.SendAddr(recipientID, []PeerAddress{addr})
			} else {
				recipient.Unlock()
			}
		}
	}
}

// selectPeersForAddrForwarding randomly selects up to maxAddrs unique peers from the provided list.
// Uses Fisher-Yates shuffle to randomly select peers without bias.
func selectPeersForAddrForwarding(peers []common.PeerId, maxAddrs int) []common.PeerId {
	if len(peers) <= maxAddrs {
		return peers
	}

	// Create a copy to avoid modifying the original slice
	shuffled := make([]common.PeerId, len(peers))
	copy(shuffled, peers)

	// Fisher-Yates shuffle for random selection
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := len(shuffled) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	// Return the first maxAddrs elements after shuffling
	return shuffled[:maxAddrs]
}
