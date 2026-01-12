package discovery

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"

	"bjoernblessin.de/go-utils/util/logger"
)

// GetAddrMsgHandler defines the interface for handling incoming getaddr messages.
// This interface is implemented in the core/domain layer and used by the infrastructure layer.
type GetAddrMsgHandler interface {
	HandleGetAddr(peerID common.PeerId)
}

// AddrMsgSender defines the interface for sending addr messages.
// This interface is implemented in the infrastructure layer and used by the core/domain layer.
type AddrMsgSender interface {
	// SendAddr sends an addr message containing a list of peer addresses to the specified peer.
	SendAddr(peerID common.PeerId, addrs []PeerAddress)
}

// GetAddrMsgSender defines the interface for sending getaddr messages.
// This interface is implemented in the infrastructure layer and used by the core/domain layer.
type GetAddrMsgSender interface {
	// SendGetAddr sends a getaddr message to the specified peer.
	// This is usually called after a successful handshake or if we run out of known peers.
	SendGetAddr(peerID common.PeerId)
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

		peer, exists := s.peerRetriever.GetPeer(pID)
		if !exists {
			continue
		}

		// The infrastructure layer will map PeerId to IP address
		// Core layer only tracks peer IDs and timestamps
		peerAddresses = append(peerAddresses, PeerAddress{
			PeerId:              pID,
			LastActiveTimestamp: peer.LastSeen,
		})
	}

	if len(peerAddresses) > 0 {
		logger.Infof("Sending addr message to peer %s with %d peer addresses", peerID, len(peerAddresses))
		go s.addrMsgSender.SendAddr(peerID, peerAddresses)
	}
}

func (s *DiscoveryService) SendGetAddr(peerID common.PeerId) {
	logger.Infof("Sending getaddr message to peer %s", peerID)
	go s.getAddrMsgSender.SendGetAddr(peerID)
}

// PeerAddress represents a peer identifier and last activity timestamp.
// The infrastructure layer handles mapping PeerId to IP addresses.
type PeerAddress struct {
	PeerId              common.PeerId
	LastActiveTimestamp int64
}
