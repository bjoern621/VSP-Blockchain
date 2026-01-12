// Package api contains domain logic for transaction operations.
package transaktion

import (
	"s3b/vsp-blockchain/rest-api/internal/common"
	adapter "s3b/vsp-blockchain/rest-api/vsgoin_node_adapter"
)

// TransaktionAPI handles transaction domain logic.
type TransaktionAPI struct {
	nodeAdapter adapter.TransactionAdapter
}

// NewTransaktionAPI creates a new TransaktionAPI with the given node adapter.
func NewTransaktionAPI(nodeAdapter adapter.TransactionAdapter) *TransaktionAPI {
	return &TransaktionAPI{
		nodeAdapter: nodeAdapter,
	}
}

// ValidationError represents a domain validation error.
type ValidationError struct {
	Message     string
	IsAuthError bool // True if the error is authentication-related (e.g., invalid private key format)
}

func (e *ValidationError) Error() string {
	return e.Message
}

// CreateTransaction validates the request and creates a transaction.
// Returns a ValidationError if validation fails, or the result from the node adapter.
func (s *TransaktionAPI) CreateTransaction(req common.TransactionRequest) (*common.TransactionResult, *ValidationError) {
	// Validate request according to arc42 validation rules (1.2 Validierungsregeln)
	if validationErr := s.validateRequest(req); validationErr != nil {
		return nil, validationErr
	}

	// Create adapter request
	adapterReq := common.TransactionRequest{
		RecipientVSAddress:  req.RecipientVSAddress,
		Amount:              uint64(req.Amount),
		SenderPrivateKeyWIF: req.SenderPrivateKeyWIF,
	}

	// Call the node adapter to create the transaction
	result, err := s.nodeAdapter.CreateTransaction(adapterReq)
	if err != nil {
		return nil, &ValidationError{
			Message:     "Internal server error",
			IsAuthError: false,
		}
	}

	return result, nil
}

// validateRequest validates the transaction request
func (s *TransaktionAPI) validateRequest(req common.TransactionRequest) *ValidationError {
	// Validate required fields
	if req.RecipientVSAddress == "" {
		return &ValidationError{
			Message:     "recipientVSAddress is required",
			IsAuthError: false,
		}
	}

	if req.SenderPrivateKeyWIF == "" {
		return &ValidationError{
			Message:     "senderPrivateKeyWIF is required",
			IsAuthError: false,
		}
	}

	// Validate amount
	if req.Amount < 1 {
		return &ValidationError{
			Message:     "amount must be at least 1",
			IsAuthError: false,
		}
	}

	// Validate private key format
	if !common.PrivateKeyWIFPattern.MatchString(req.SenderPrivateKeyWIF) {
		return &ValidationError{
			Message:     "Invalid private key format",
			IsAuthError: true,
		}
	}

	// Validate recipient address format
	if !common.VsAddressPattern.MatchString(req.RecipientVSAddress) {
		return &ValidationError{
			Message:     "Invalid recipient address format",
			IsAuthError: false,
		}
	}

	return nil
}
