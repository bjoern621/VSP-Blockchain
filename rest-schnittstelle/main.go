package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	sw "s3b/vsp-blockchain/rest-api/api_adapter"
	"strings"

	"bjoernblessin.de/go-utils/util/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/gin-gonic/gin"
)

//go:embed swagger-ui/*
var swaggerContent embed.FS

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

	routes := sw.ApiHandleFunctions{}

	log.Printf("Server started")

	router := sw.NewRouter(routes)

	fsys, _ := fs.Sub(swaggerContent, "swagger-ui")
	router.StaticFS("/swagger-ui/", http.FS(fsys))

	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "./swagger-ui")
	})

	log.Fatal(router.Run(":8080"))

}
