package core

import (
	"net/netip"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"
)

// FullNetworkInfo contains all available information about a peer.
type FullNetworkInfo struct {
	PeerID            peer.PeerID
	ListeningEndpoint netip.AddrPort
	InboundAddresses  []netip.AddrPort
	HasOutboundConn   bool
}

// NetworkInfoProvider provides access to network-level information about peers.
type NetworkInfoProvider interface {
	// GetAllNetworkInfo returns all available information for all peers.
	GetAllNetworkInfo() []FullNetworkInfo
}
