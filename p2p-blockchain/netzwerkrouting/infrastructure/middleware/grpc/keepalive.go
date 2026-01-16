package grpc

import (
	"context"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"

	"bjoernblessin.de/go-utils/util/logger"
	"google.golang.org/protobuf/types/known/emptypb"
)

// HeartbeatBing handles incoming HeartbeatBing messages from peers.
// This method is called when another peer sends a bing to check liveness.
// It updates the peer's LastSeen timestamp and responds with HeartbeatBong.
func (s *Server) HeartbeatBing(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	peerId := s.GetPeerId(ctx)
	logger.Debugf("Received HeartbeatBing from peer %s", peerId)

	s.keepaliveService.HandleHeartbeatBing(peerId)

	return &emptypb.Empty{}, nil
}

// HeartbeatBong handles incoming HeartbeatBong messages from peers.
// This method is called when another peer responds to our HeartbeatBing.
// It updates the peer's LastSeen timestamp.
func (s *Server) HeartbeatBong(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	peerId := s.GetPeerId(ctx)
	logger.Debugf("Received HeartbeatBong from peer %s", peerId)

	s.keepaliveService.HandleHeartbeatBong(peerId)

	return &emptypb.Empty{}, nil
}

// SendHeartbeatBing sends a HeartbeatBing message to the specified peer and waits for a HeartbeatBong response.
// This method is used as a keep-alive mechanism to check if a peer is still responsive.
func (c *Client) SendHeartbeatBing(peerID common.PeerId) {
	SendHelper(c, peerID, "HeartbeatBing", pb.NewPeerDiscoveryClient, func(client pb.PeerDiscoveryClient) error {
		_, err := client.HeartbeatBing(context.Background(), &emptypb.Empty{})
		return err
	})
}

// SendHeartbeatBong sends a HeartbeatBong message to the specified peer.
// This method is used to respond to a HeartbeatBing from another peer.
func (c *Client) SendHeartbeatBong(peerID common.PeerId) {
	SendHelper(c, peerID, "HeartbeatBong", pb.NewPeerDiscoveryClient, func(client pb.PeerDiscoveryClient) error {
		_, err := client.HeartbeatBong(context.Background(), &emptypb.Empty{})
		return err
	})
}
