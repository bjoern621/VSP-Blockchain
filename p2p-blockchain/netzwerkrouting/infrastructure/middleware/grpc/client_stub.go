package grpc

import (
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/handshake"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/infrastructure/middleware/grpc/networkinfo"
)

// Client represents the P2P gRPC client for peer-to-peer communication.
// This is the application stub / grpc proxy.
// It contains no domain logic, only type transformation and delegation.
type Client struct {
	networkInfoRegistry *networkinfo.NetworkInfoRegistry
}

// Compile-time check that Client implements HandshakeInitiator
var _ handshake.HandshakeInitiator = (*Client)(nil)

func NewClient(networkInfoRegistry *networkinfo.NetworkInfoRegistry) *Client {
	return &Client{
		networkInfoRegistry: networkInfoRegistry,
	}
}
