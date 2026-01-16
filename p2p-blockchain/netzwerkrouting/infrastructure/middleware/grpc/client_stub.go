package grpc

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/infrastructure/middleware/grpc/networkinfo"

	"bjoernblessin.de/go-utils/util/logger"
	"google.golang.org/grpc"
)

// Client represents the P2P gRPC client for peer-to-peer communication.
// This is the application stub / grpc proxy.
// It contains no domain logic, only type transformation and delegation.
type Client struct {
	networkInfoRegistry *networkinfo.NetworkInfoRegistry
}

func NewClient(networkInfoRegistry *networkinfo.NetworkInfoRegistry) *Client {
	return &Client{
		networkInfoRegistry: networkInfoRegistry,
	}
}

// SendHelper is a generic helper to send gRPC messages to a peer.
// It will retrieve peer connection, create specific gRPC client, and handle calling the grpc method.
// Should generally be used to implement SendXXX methods on Client.
func SendHelper[T any](c *Client, peerID common.PeerId, method string, newClient func(grpc.ClientConnInterface) T, send func(T) error) {
	conn, ok := c.networkInfoRegistry.GetConnection(peerID)
	if !ok {
		logger.Warnf("failed to send %s: no connection for peer %s", method, peerID)
		return
	}

	client := newClient(conn)
	if err := send(client); err != nil {
		logger.Warnf("failed to send %s to peer %s: %v", method, peerID, err)
	}
}
