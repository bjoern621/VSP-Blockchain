package handshake

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
)

// handshakeService implements HandshakeMsgHandler (for infrastructure), HandshakeInitiator (for api),
// and ObservableHandshakeService (for connection observers) with the actual domain logic.
type handshakeService struct {
	handshakeMsgSender HandshakeMsgSender
	peerRetriever      peerRetriever
	errorMsgSender     errorMsgSender
	*observerManager
}

func NewHandshakeService(handshakeMsgSender HandshakeMsgSender, peerRetriever peerRetriever, errorMsgSender errorMsgSender) *handshakeService {
	return &handshakeService{
		handshakeMsgSender: handshakeMsgSender,
		peerRetriever:      peerRetriever,
		errorMsgSender:     errorMsgSender,
		observerManager:    newObserverManager(),
	}
}

// peerRetriever is an interface for retrieving peers.
// It is implemented by peer.PeerStore.
type peerRetriever interface {
	GetPeer(id common.PeerId) (*common.Peer, bool)
}
