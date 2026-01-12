package discovery

import (
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/data/peer"
)

// DiscoveryService provides peer discovery functionality.
// This includes (1) querying a registry for peers, (2) asking neighbors for their known peers as well as (3) keeping track of active peers through heartbeats.
type DiscoveryService struct {
	querier          RegistryQuerier
	peerCreator      peer.PeerCreator
	addrMsgSender    AddrMsgSender
	peerRetriever    DiscoveryPeerRetriever
	getAddrMsgSender GetAddrMsgSender
}

// NewDiscoveryService creates a new DiscoveryService.
func NewDiscoveryService(querier RegistryQuerier, peerCreator peer.PeerCreator, addrMsgSender AddrMsgSender, peerRetriever DiscoveryPeerRetriever, getAddrMsgSender GetAddrMsgSender) *DiscoveryService {
	return &DiscoveryService{
		querier:          querier,
		peerCreator:      peerCreator,
		addrMsgSender:    addrMsgSender,
		peerRetriever:    peerRetriever,
		getAddrMsgSender: getAddrMsgSender,
	}
}
