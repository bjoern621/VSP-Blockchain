package handshake

import (
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/data/peer"
)

// handshakeService implements HandshakeMsgHandler (for infrastructure) and HandshakeInitiator (for api) with the actual domain logic.
type handshakeService struct {
	handshakeMsgSender HandshakeMsgSender
	peerRetriever      peer.PeerRetriever
}

func NewHandshakeService(handshakeMsgSender HandshakeMsgSender, peerRetriever peer.PeerRetriever) *handshakeService {
	return &handshakeService{
		handshakeMsgSender: handshakeMsgSender,
		peerRetriever:      peerRetriever,
	}
}
