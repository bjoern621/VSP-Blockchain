package discovery

import "s3b/vsp-blockchain/p2p-blockchain/internal/common"

// errorMsgSender defines the interface for sending error/reject messages to peers.
// This allows the core layer to send reject messages without depending on the API layer.
type errorMsgSender interface {
	// SendReject sends a reject message to the specified peer
	SendReject(peerId common.PeerId, errorType int32, rejectedMessageType string, data []byte)
}

// DiscoveryService provides peer discovery functionality.
// This includes (1) querying a registry for peers and (2) asking neighbors for their known peers.
type DiscoveryService struct {
	querier          RegistryQuerier
	addrMsgSender    AddrMsgSender
	peerRetriever    peerRetriever
	getAddrMsgSender GetAddrMsgSender
	errorMsgSender   errorMsgSender
}

// NewDiscoveryService creates a new DiscoveryService.
func NewDiscoveryService(
	querier RegistryQuerier,
	addrMsgSender AddrMsgSender,
	peerRetriever peerRetriever,
	getAddrMsgSender GetAddrMsgSender,
	errorMsgSender errorMsgSender,
) *DiscoveryService {
	return &DiscoveryService{
		querier:          querier,
		addrMsgSender:    addrMsgSender,
		peerRetriever:    peerRetriever,
		getAddrMsgSender: getAddrMsgSender,
		errorMsgSender:   errorMsgSender,
	}
}
