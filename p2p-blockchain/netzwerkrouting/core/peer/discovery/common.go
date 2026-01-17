package discovery

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/data/peer"
)

// peerRetriever is an interface for retrieving peers specifically for discovery purposes.
// Implemented by the data layer's PeerStore.
type peerRetriever interface {
	// GetAllPeers retrieves all known peers.
	GetAllPeers() []common.PeerId
	// GetPeer retrieves a peer by its ID.
	GetPeer(id common.PeerId) (*peer.Peer, bool)
	// GetAllConnectedPeers retrieves all connected peers' IDs (both inbound and outbound).
	// All peers with StateConnected are considered connected.
	GetAllConnectedPeers() []common.PeerId
}
