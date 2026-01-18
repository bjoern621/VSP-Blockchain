package grpc

import (
	"context"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"

	"bjoernblessin.de/go-utils/util/logger"
	"google.golang.org/protobuf/types/known/emptypb"
)

// SendReject sends a reject message to the specified peer.
// This is used to signal errors in received messages back to the sender.
func (c *Client) SendReject(peerID common.PeerId, errorType int32, rejectedMessageType string, data []byte) {
	SendHelper(c, peerID, "Reject", pb.NewErrorHandlingClient, func(client pb.ErrorHandlingClient) error {
		_, err := client.Reject(context.Background(), &pb.Error{
			ErrorType:           pb.ErrorType(errorType),
			RejectedMessageType: rejectedMessageType,
			Data:                data,
		})
		return err
	})
}

// This is called when a peer signals a fault in a message sent by this node.
// The error is logged for debugging and monitoring purposes.
func (s *Server) Reject(ctx context.Context, req *pb.Error) (*emptypb.Empty, error) {
	peerID := s.GetPeerId(ctx)

	logger.Infof("[error_handling] Reject message received from peer %s: error_type=%v, message_type=%s",
		peerID, req.ErrorType, req.RejectedMessageType)

	return &emptypb.Empty{}, nil
}
