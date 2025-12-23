// Package handlers contains HTTP handlers for the REST api endpoints.
package api

import (
	"encoding/json"
	"net/http"
	"regexp"
	"s3b/vsp-blockchain/rest-api/adapter/api"
	"s3b/vsp-blockchain/rest-api/adapter/core"
	"s3b/vsp-blockchain/rest-api/internal/common"

	"bjoernblessin.de/go-utils/util/logger"
)

// TransactionRequest represents the JSON request body for POST /transaction.
// Matches the OpenAPI specification schema.
type TransactionRequest struct {
	RecipientVSAddress  string `json:"recipientVSAddress"`
	Amount              int64  `json:"amount"`
	SenderPrivateKeyWIF string `json:"senderPrivateKeyWIF"`
}

// TransactionHandler handles POST /transaction requests.
type TransactionHandler struct {
	nodeAdapter api.TransactionAdapterAPI
}

// NewTransactionHandler creates a new TransactionHandler with the given node adapter.
func NewTransactionHandler(nodeAdapter api.TransactionAdapterAPI) *TransactionHandler {
	return &TransactionHandler{
		nodeAdapter: nodeAdapter,
	}
}

// Validation patterns as specified in OpenAPI and arc42 documentation.
// Base58Check format: starts with 5, followed by 50 Base58 characters (WIF format).
var privateKeyWIFPattern = regexp.MustCompile(`^5[1-9A-HJ-NP-Za-km-z]{50}$`)

// VSAddress validation: Base58Check encoded (starts with 1 for mainnet addresses).
var vsAddressPattern = regexp.MustCompile(`^1[1-9A-HJ-NP-Za-km-z]{25,34}$`)

// ServeHTTP handles the HTTP request for transaction creation.
func (h *TransactionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Validate HTTP method
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warnf("Failed to decode transaction request: %v", err)
		http.Error(w, "Invalid JSON request body", http.StatusBadRequest)
		return
	}

	// Validate request according to arc42 validation rules (1.2 Validierungsregeln)
	if validationErr := h.validateRequest(req); validationErr != nil {
		logger.Warnf("Transaction request validation failed: %v", validationErr)
		http.Error(w, validationErr.Error(), validationErr.statusCode)
		return
	}

	// Create adapter request
	adapterReq := core.TransactionRequest{
		RecipientVSAddress:  req.RecipientVSAddress,
		Amount:              uint64(req.Amount),
		SenderPrivateKeyWIF: req.SenderPrivateKeyWIF,
	}

	// Call the node adapter to create the transaction
	result, err := h.nodeAdapter.CreateTransaction(r.Context(), adapterReq)
	if err != nil {
		logger.Errorf("Failed to create transaction via adapter: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Handle result based on error code
	h.writeResponse(w, result)
}

// ValidationError represents a validation error with HTTP status code.
type ValidationError struct {
	message    string
	statusCode int
}

func (e *ValidationError) Error() string {
	return e.message
}

// validateRequest validates the transaction request
func (h *TransactionHandler) validateRequest(req TransactionRequest) *ValidationError {
	// Validate required fields
	if req.RecipientVSAddress == "" {
		return &ValidationError{
			message:    "recipientVSAddress is required",
			statusCode: http.StatusBadRequest,
		}
	}

	if req.SenderPrivateKeyWIF == "" {
		return &ValidationError{
			message:    "senderPrivateKeyWIF is required",
			statusCode: http.StatusBadRequest,
		}
	}

	// Validate amount
	if req.Amount < 1 {
		return &ValidationError{
			message:    "amount must be at least 1",
			statusCode: http.StatusBadRequest,
		}
	}

	// Validate private key format
	if !privateKeyWIFPattern.MatchString(req.SenderPrivateKeyWIF) {
		return &ValidationError{
			message:    "Invalid private key format",
			statusCode: http.StatusUnauthorized,
		}
	}

	// Validate recipient address format
	if !vsAddressPattern.MatchString(req.RecipientVSAddress) {
		return &ValidationError{
			message:    "Invalid recipient address format",
			statusCode: http.StatusBadRequest,
		}
	}

	return nil
}

// writeResponse writes the appropriate HTTP response based on the transaction result.
func (h *TransactionHandler) writeResponse(w http.ResponseWriter, result *core.TransactionResult) {
	if result.Success {
		// 201 Created - Transaction successfully executed
		w.WriteHeader(http.StatusCreated)
		logger.Infof("Transaction created successfully: %s", result.TransactionID)
		return
	}

	switch result.ErrorCode {
	case common.ErrorCodeInvalidPrivateKey:
		// 401 Unauthorized - Invalid private key
		http.Error(w, result.ErrorMessage, http.StatusUnauthorized)
	case common.ErrorCodeInsufficientFunds:
		// 400 Bad Request - Insufficient funds
		http.Error(w, result.ErrorMessage, http.StatusBadRequest)
	case common.ErrorCodeValidationFailed, common.ErrorCodeBroadcastFailed:
		// 400 Bad Request - Validation or broadcast failed
		http.Error(w, result.ErrorMessage, http.StatusBadRequest)
	default:
		// 500 Internal Server Error - Unexpected error
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}

	logger.Warnf("Transaction failed: code=%d, message=%s", result.ErrorCode, result.ErrorMessage)
}
