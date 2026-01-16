package grpc

import (
	"context"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"

	"bjoernblessin.de/go-utils/util/logger"
	"google.golang.org/protobuf/types/known/emptypb"
)

// HeartbeatPing handles incoming HeartbeatPing messages from peers.
// This method is called when another peer sends a ping to check liveness.
// It updates the peer's LastSeen timestamp and responds with HeartbeatPong.
func (s *Server) HeartbeatPing(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	peerId := s.GetPeerId(ctx)
	logger.Debugf("Received HeartbeatPing from peer %s", peerId)

	s.keepaliveService.HandleHeartbeatPing(peerId)

	return &emptypb.Empty{}, nil
}

// HeartbeatPong handles incoming HeartbeatPong messages from peers.
// This method is called when another peer responds to our HeartbeatPing.
// It updates the peer's LastSeen timestamp.
func (s *Server) HeartbeatPong(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	peerId := s.GetPeerId(ctx)
	logger.Debugf("Received HeartbeatPong from peer %s", peerId)

	s.keepaliveService.HandleHeartbeatPong(peerId)

	return &emptypb.Empty{}, nil
}

// SendHeartbeatPing sends a HeartbeatPing message to the specified peer and waits for a HeartbeatPong response.
// This method is used as a keep-alive mechanism to check if a peer is still responsive.
func (c *Client) SendHeartbeatPing(peerID common.PeerId) {
	SendHelper(c, peerID, "HeartbeatPing", pb.NewPeerDiscoveryClient, func(client pb.PeerDiscoveryClient) error {
		_, err := client.HeartbeatPing(context.Background(), &emptypb.Empty{})
		return err
	})
}

// SendHeartbeatPong sends a HeartbeatPong message to the specified peer.
// This method is used to respond to a HeartbeatPing from another peer.
func (c *Client) SendHeartbeatPong(peerID common.PeerId) {
	SendHelper(c, peerID, "HeartbeatPong", pb.NewPeerDiscoveryClient, func(client pb.PeerDiscoveryClient) error {
		_, err := client.HeartbeatPong(context.Background(), &emptypb.Empty{})
		return err
	})
}
