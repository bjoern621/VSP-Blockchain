package grpc

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/infrastructure/middleware/grpc/networkinfo"

	"bjoernblessin.de/go-utils/util/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// peerDisconnector defines an interface for disconnecting peers.
// Implemented by the core layer.
type peerDisconnector interface {
	Disconnect(peerID common.PeerId) error
}

// Client represents the P2P gRPC client for peer-to-peer communication.
// This is the application stub / grpc proxy.
// It contains no domain logic, only type transformation and delegation.
type Client struct {
	networkInfoRegistry *networkinfo.NetworkInfoRegistry
	peerDisconnector    peerDisconnector
}

func NewClient(networkInfoRegistry *networkinfo.NetworkInfoRegistry, peerDisconnector peerDisconnector) *Client {
	return &Client{
		networkInfoRegistry: networkInfoRegistry,
		peerDisconnector:    peerDisconnector,
	}
}

// SendHelper is a generic helper to send gRPC messages to a peer.
// It will retrieve peer connection, create specific gRPC client, and handle calling the grpc method.
// Should generally be used to implement SendXXX methods on Client.
//
// Usage example:
//
//	SendHelper(c, peerID, "Ack", pb.NewConnectionEstablishmentClient, func(client pb.ConnectionEstablishmentClient) error {
//		_, err := client.Ack(context.Background(), &emptypb.Empty{})
//		return err
//	})
func SendHelper[T any](c *Client, peerID common.PeerId, method string, newClient func(grpc.ClientConnInterface) T, send func(T) error) {
	conn, ok := c.networkInfoRegistry.GetConnection(peerID)
	if !ok {
		logger.Warnf("[client_stub] failed to send %s: no connection for peer %s", method, peerID)
		return
	}

	client := newClient(conn)
	if err := send(client); err != nil {
		logger.Warnf("[client_stub] failed to send %s to peer %s: %v", method, peerID, err)

		// Check if this is a fatal connection error that indicates the peer is unreachable
		// This is just a best-effort attempt to keep the peer store clean
		if isFatalConnectionError(err) {
			logger.Infof("[client_stub] Fatal connection error detected for peer %s, triggering disconnect", peerID)
			_ = c.peerDisconnector.Disconnect(peerID)
		}
	}
}

// isFatalConnectionError checks if a gRPC error indicates a permanent connection failure.
func isFatalConnectionError(err error) bool {
	st, ok := status.FromError(err)
	if !ok {
		return false
	}

	// These error codes indicate the peer is unreachable or the connection is dead
	switch st.Code() {
	case codes.Unavailable: // Connection failed, timeout, or refused
		return true
	case codes.DeadlineExceeded: // Request timeout
		return true
	default:
		return false
	}
}
