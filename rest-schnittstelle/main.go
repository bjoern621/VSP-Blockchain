package main

import (
	"context"
	"net/http"

	"s3b/vsp-blockchain/rest-schnittstelle/internal/api/handlers"
	"s3b/vsp-blockchain/rest-schnittstelle/internal/api/middleware"

	"bjoernblessin.de/go-utils/util/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	testpb "s3b/vsp-blockchain/p2p-blockchain/proto/pb"
)

func main() {
	logger.Infof("Running...")

	// gRPC Client Setup

	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Errorf("failed to create connection to gRPC server: %v", err)
	}
	defer conn.Close()

	client := testpb.NewTestClient(conn)
	resp, err := client.TestRPC(context.Background(), &testpb.TestRequest{Message: "Hello from REST Schnittstelle"})
	if err != nil {
		logger.Warnf("gRPC call failed: %v", err)
	} else {
		logger.Infof("gRPC response: %s", resp.Message)
	}

	// REST API Server

	mux := http.NewServeMux()
	mux.HandleFunc("GET /test", handlers.TestHandler)

	handler := middleware.Logging(mux)

	err = http.ListenAndServe(":8080", handler)
	logger.Errorf("%v", err)
}
