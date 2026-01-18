package handshake

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
)

// handshakeService implements HandshakeMsgHandler (for infrastructure) and HandshakeInitiator (for api) with the actual domain logic.
type handshakeService struct {
	handshakeMsgSender HandshakeMsgSender
	peerRetriever      peerRetriever
}

func NewHandshakeService(handshakeMsgSender HandshakeMsgSender, peerRetriever peerRetriever) *handshakeService {
	return &handshakeService{
		handshakeMsgSender: handshakeMsgSender,
		peerRetriever:      peerRetriever,
	}
}

// peerRetriever is an interface for retrieving peers.
// It is implemented by peer.PeerStore.
type peerRetriever interface {
	GetPeer(id common.PeerId) (*common.Peer, bool)
}
