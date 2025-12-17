package main

import (
	"net/http"
	"os"
	"strings"

	"s3b/vsp-blockchain/rest-api/internal/api/handlers"
	"s3b/vsp-blockchain/rest-api/internal/api/middleware"

	"bjoernblessin.de/go-utils/util/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	logger.Infof("Running...")

	// gRPC Client Setup

	grpcAddrPort := strings.TrimSpace(os.Getenv("APP_GRPC_ADDR_PORT"))
	if grpcAddrPort == "" {
		grpcAddrPort = "localhost:50050"
	}

	conn, err := grpc.NewClient(grpcAddrPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Errorf("failed to create connection to gRPC server: %v", err)
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			logger.Errorf("failed to close gRPC connection: %v", err)
		}
	}(conn)

	// REST API Server

	mux := http.NewServeMux()
	mux.HandleFunc("GET /test", handlers.TestHandler)

	handler := middleware.Logging(mux)

	err = http.ListenAndServe(":8080", handler)
	logger.Errorf("%v", err)
}
