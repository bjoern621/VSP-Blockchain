package grpc

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/handshake"
)

// Client represents the P2P gRPC client for peer-to-peer communication.
// This is the application stub / grpc proxy.
// It contains no domain logic, only type transformation and delegation.
type Client struct {
	grpcClient   pb.ConnectionEstablishmentClient
	peerRegistry *PeerRegistry
}

// Compile-time check that Client implements HandshakeInitiator
var _ handshake.HandshakeInitiator = (*Client)(nil)

func NewClient(peerRegistry *PeerRegistry) *Client {
	return &Client{
		peerRegistry: peerRegistry,
	}
}
