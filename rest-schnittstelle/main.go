package main

import (
	"net/http"
	"s3b/vsp-blockchain/rest-schnittstelle/internal/api/handlers"
	"s3b/vsp-blockchain/rest-schnittstelle/internal/api/middleware"

	"bjoernblessin.de/go-utils/util/logger"
)

func main() {
	logger.Infof("Running...")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /test", handlers.TestHandler)

	handler := middleware.Logging(mux)

	err := http.ListenAndServe(":8080", handler)
	logger.Errorf("%v", err)
}
