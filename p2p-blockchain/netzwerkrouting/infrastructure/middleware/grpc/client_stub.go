package grpc

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"
)

type Client struct {
	grpcClient   pb.ConnectionEstablishmentClient
	peerRegistry *PeerRegistry
}

func NewClient(peerRegistry *PeerRegistry) *Client {
	return &Client{
		peerRegistry: peerRegistry,
	}
}
