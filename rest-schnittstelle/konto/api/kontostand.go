// Package handler contains HTTP handlers for the konto component.
package api

import (
	"encoding/json"
	"net/http"
	"s3b/vsp-blockchain/rest-api/adapter/api"
	"s3b/vsp-blockchain/rest-api/internal/common"

	"bjoernblessin.de/go-utils/util/logger"
)

// KontostandResponse represents the JSON response for GET /balance.
type KontostandResponse struct {
	Kontostand uint64 `json:"balance"`
}

// KontostandHandler handles GET /balance requests.
type KontostandHandler struct {
	kontoAdapter api.KontoAdapterAPI
}

// NewKontostandHandler creates a new KontostandHandler with the given adapter.
func NewKontostandHandler(kontoAdapter api.KontoAdapterAPI) *KontostandHandler {
	return &KontostandHandler{
		kontoAdapter: kontoAdapter,
	}
}

// ServeHTTP handles the HTTP request for Kontostand queries.
func (h *KontostandHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get VSAddress from query parameter
	vsAddress := r.URL.Query().Get("VSAddress")
	if !h.validateAddress(w, vsAddress) {
		return
	}

	// Query assets via adapter
	result, err := h.kontoAdapter.GetAssets(r.Context(), vsAddress)
	if err != nil {
		logger.Errorf("Failed to get assets: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !result.Success {
		logger.Warnf("Assets query failed: %s", result.ErrorMessage)
		http.Error(w, result.ErrorMessage, http.StatusBadRequest)
		return
	}

	h.handleSuccess(w, result)
}

func (h *KontostandHandler) handleSuccess(w http.ResponseWriter, result *common.AssetsResult) {
	var kontostand uint64
	for _, asset := range result.Assets {
		kontostand += asset.Value
	}

	response := KontostandResponse{
		Kontostand: kontostand,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Errorf("Failed to encode Kontostand response: %v", err)
	}
}

func (h *KontostandHandler) validateAddress(w http.ResponseWriter, vsAddress string) bool {
	if vsAddress == "" {
		http.Error(w, "VSAddress query parameter is required", http.StatusBadRequest)
		return false
	}

	// Validate VSAddress format
	if !common.VsAddressPattern.MatchString(vsAddress) {
		http.Error(w, "Invalid VSAddress format", http.StatusBadRequest)
		return false
	}
	return true
}
