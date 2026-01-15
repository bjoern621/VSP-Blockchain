package discovery

import (
	"math/rand"
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

	// Update last seen timestamps for all received addresses
	for _, addr := range addrs {
		peer, exists := s.peerRetriever.GetPeer(addr.PeerId)
		assert.Assert(exists, "peer should already be registered by infrastructure layer")

		if peer.LastSeen < addr.LastActiveTimestamp {
			peer.Lock()
			peer.LastSeen = addr.LastActiveTimestamp
			peer.Unlock()
		}

		logger.Infof("Discovered peer from Addr msg: PeerId=%s, LastSeen=%v",
			addr.PeerId, time.Unix(addr.LastActiveTimestamp, 0))
	}

	// Forward addresses to random neighbors (per-recipient filtering happens in forwardAddrs)
	if len(addrs) > 0 {
		s.forwardAddrs(addrs, peerID)
	}
}

// forwardAddrs forwards the given addrs to neighboring peers.
// Forwarding rules:
//   - Do not forward to the peer from which we received the addr
//   - Do not forward an address to a peer that has already received it
//   - Forward to a small random selection of connected peers (1-3 peers, average of 2)
func (s *DiscoveryService) forwardAddrs(addrs []PeerAddress, sender common.PeerId) {
	// Get all connected peers (both inbound and outbound)
	connectedPeers := s.peerRetriever.GetAllConnectedPeers()

	// Build set of address peer IDs to exclude
	addrPeerIds := make(map[common.PeerId]bool)
	for _, addr := range addrs {
		addrPeerIds[addr.PeerId] = true
	}

	// Filter out: 1) sender, 2) peers mentioned in the addr list
	eligiblePeers := make([]common.PeerId, 0, len(connectedPeers))
	for _, peerID := range connectedPeers {
		if peerID == sender {
			continue // Don't forward to sender
		}
		if addrPeerIds[peerID] {
			continue // Don't forward to peers mentioned in the addr list
		}
		eligiblePeers = append(eligiblePeers, peerID)
	}

	if len(eligiblePeers) == 0 {
		logger.Debugf("No eligible peers for addr forwarding")
		return
	}

	// Select 1-3 random peers (average of 2) for forwarding
	numPeers := 2 // Average of 2 as per issue #75
	if len(eligiblePeers) < numPeers {
		numPeers = len(eligiblePeers)
	}
	selectedPeers := selectPeersForAddrForwarding(eligiblePeers, numPeers)

	// Forward addresses to selected peers, filtering per-recipient
	for _, recipientID := range selectedPeers {
		recipient, exists := s.peerRetriever.GetPeer(recipientID)
		if !exists {
			continue
		}

		// Filter addresses that haven't been sent to this recipient yet
		recipient.Lock()
		if recipient.AddrsSentTo == nil {
			recipient.AddrsSentTo = make(map[common.PeerId]bool)
		}

		filteredAddrs := make([]PeerAddress, 0, len(addrs))
		for _, addr := range addrs {
			if !recipient.AddrsSentTo[addr.PeerId] {
				filteredAddrs = append(filteredAddrs, addr)
				recipient.AddrsSentTo[addr.PeerId] = true
			}
		}
		recipient.Unlock()

		if len(filteredAddrs) > 0 {
			logger.Infof("Forwarding %d addresses to peer %s", len(filteredAddrs), recipientID)
			s.addrMsgSender.SendAddr(recipientID, filteredAddrs)
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
