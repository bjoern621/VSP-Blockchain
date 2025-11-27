package main

import (
	"s3b/vsp-blockchain/p2p-blockchain/app/core"
	"strconv"

	"bjoernblessin.de/go-utils/util/env"
	"bjoernblessin.de/go-utils/util/logger"
)

// type server struct {
// 	pb.UnimplementedTestServer
// }

// func (s *server) TestRPC(ctx context.Context, req *pb.TestRequest) (*pb.TestResponse, error) {
// 	logger.Infof("Received: %s", req.Message)

// 	return &pb.TestResponse{
// 		Message: fmt.Sprintf("Echo: %s", req.Message),
// 	}, nil
// }

const defaultPort = 50051
const portEnvVar = "PORT"

func getPortFromEnv() uint16 {
	portStr := env.ReadNonEmptyRequiredEnv(portEnvVar)

	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		logger.Errorf("invalid PORT value: %s, must be between 0 and 65535", portStr)
	}

	if port == 0 {
		port = defaultPort
	}

	return uint16(port)
}

func main() {
	logger.Infof("Running...")

	port := getPortFromEnv()
	logger.Infof("Running on port %d...", port)

	core.NewServer().Start(port)

	select {}

	// ctx := context.Background()
	// ip := netip.MustParseAddr("::1")

	// if port == 50051 {
	// 	err := core.ConnectTo(ctx, ip, 50052)
	// 	if err != nil {
	// 		logger.Errorf("connection failed: %v", err)
	// 	}
	// }

	// listener, err := net.Listen("tcp", ":50051")
	// if err != nil {
	// 	logger.Errorf("failed to listen: %v", err)
	// }

	// grpcServer := grpc.NewServer()
	// pb.RegisterTestServer(grpcServer, &server{})

	// logger.Infof("gRPC server listening on :50051")

	// if err := grpcServer.Serve(listener); err != nil {
	// 	logger.Errorf("failed to serve: %v", err)
	// }
}
