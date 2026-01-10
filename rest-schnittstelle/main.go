package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	sw "s3b/vsp-blockchain/rest-api/api_adapter"
	"s3b/vsp-blockchain/rest-api/internal/pb"
	"s3b/vsp-blockchain/rest-api/konto"
	transactionapi "s3b/vsp-blockchain/rest-api/transaktion"
	"s3b/vsp-blockchain/rest-api/vsgoin_node_adapter"
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

	// Dependencies

	appServiceClient := pb.NewAppServiceClient(conn)
	transactionAdapter := vsgoin_node_adapter.NewTransactionAdapterImpl(appServiceClient)
	kontoAdapter := vsgoin_node_adapter.NewKontoAdapter(conn)
	kontostand := konto.NewKeyGeneratorImpl(transactionAdapter)
	transactionApi := transactionapi.NewTransaktionAPI(transactionAdapter)
	kontostandService := konto.NewKontostandService(kontoAdapter)

	// REST API Server
	routes := sw.ApiHandleFunctions{
		KeyToolsAPI: *sw.NewKeyToolsAPI(kontostand),
		PaymentAPI:  *sw.NewPaymentAPI(transactionApi, kontostandService),
	}

	log.Printf("Server started")

	router := sw.NewRouter(routes)

	fsys, _ := fs.Sub(swaggerContent, "swagger-ui")
	router.StaticFS("/swagger-ui/", http.FS(fsys))

	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "./swagger-ui")
	})

	log.Fatal(router.Run(":8080"))

}
