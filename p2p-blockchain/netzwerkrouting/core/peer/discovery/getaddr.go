package discovery

import (
	"time"

	"s3b/vsp-blockchain/p2p-blockchain/internal/common"

	"bjoernblessin.de/go-utils/util/logger"
)

// GetAddrMsgHandler defines the interface for handling incoming getaddr messages.
// This interface is implemented in the core/domain layer and used by the infrastructure layer.
type GetAddrMsgHandler interface {
	HandleGetAddr(peerID common.PeerId)
}

func (s *DiscoveryService) HandleGetAddr(peerID common.PeerId) {
	// Read all known peers and send them back in an addr message
	peers := s.peerRetriever.GetAllPeers()
	peerAddresses := make([]PeerAddress, 0, len(peers))

	for _, pID := range peers {
		if pID == peerID {
			// Don't send the requesting peer's own address back
			continue
		}

		_, exists := s.peerRetriever.GetPeer(pID)
		if !exists {
			continue
		}

		// The infrastructure layer will map PeerId to IP address
		// Core layer only tracks peer IDs and timestamps
		peerAddresses = append(peerAddresses, PeerAddress{
			PeerId:              pID,
			LastActiveTimestamp: time.Now().Unix(), // TODO anpassen wenn heartbeat fertig
		})
	}

	if len(peerAddresses) > 0 {
		logger.Infof("Sending addr message to peer %s with %d peer addresses", peerID, len(peerAddresses))
		go s.addrMsgSender.SendAddr(peerID, peerAddresses)
	}
}

// GetAddrMsgSender defines the interface for sending getaddr messages.
type GetAddrMsgSender interface {
	// SendGetAddr sends a getaddr message to the specified peer.
	// This is usually called after a successful handshake.
	SendGetAddr(peerID common.PeerId)
}

func (s *DiscoveryService) SendGetAddr(peerID common.PeerId) {
	logger.Infof("Sending getaddr message to peer %s", peerID)
	// Implementation for sending getaddr messages goes here.
}

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
		// The infrastructure layer will handle the IP address mapping
		logger.Debugf("New peer discovered: PeerId=%s, LastSeen=%v",
			addr.PeerId, time.Unix(addr.LastActiveTimestamp, 0))
		// TODO: Trigger peer creation and connection for new peer
	}
}

// PeerAddress represents a peer identifier and last activity timestamp.
// The infrastructure layer handles mapping PeerId to IP addresses.
type PeerAddress struct {
	PeerId              common.PeerId
	LastActiveTimestamp int64
}
