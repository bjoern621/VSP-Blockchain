package main

import (
	"context"
	"fmt"
	"net"

	testpb "s3b/vsp-blockchain/p2p-blockchain/proto/pb"

	"bjoernblessin.de/go-utils/util/logger"
	"google.golang.org/grpc"
)

type server struct {
	testpb.UnimplementedTestServer
}

func (s *server) TestRPC(ctx context.Context, req *testpb.TestRequest) (*testpb.TestResponse, error) {
	logger.Infof("Received: %s", req.Message)

	return &testpb.TestResponse{
		Message: fmt.Sprintf("Echo: %s", req.Message),
	}, nil
}

func main() {
	logger.Infof("Running...")

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Errorf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	testpb.RegisterTestServer(grpcServer, &server{})

	logger.Infof("gRPC server listening on :50051")

	if err := grpcServer.Serve(listener); err != nil {
		logger.Errorf("failed to serve: %v", err)
	}
}
