package main

import (
	"net/http"
	"s3b/vsp-blockchain/rest-schnittstelle/internal/api/handlers"

	"bjoernblessin.de/go-utils/util/logger"
)

func main() {
	logger.Infof("Running...")

	http.HandleFunc("/test", handlers.TestHandler)
	logger.Errorf("http server failed: %v", http.ListenAndServe(":8080", nil))
}
