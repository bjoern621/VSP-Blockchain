// Package adapter implements the V$Goin-Node-Adapter pattern as specified in the arc42 documentation.
// It encapsulates all interactions with the local V$Goin SPV Node, providing a stable internal interface
// and decoupling the REST api from gRPC implementation details.
package core

import (
	"context"
	"fmt"
	"s3b/vsp-blockchain/rest-api/internal/common"
	"time"

	"s3b/vsp-blockchain/rest-api/internal/pb"

	"google.golang.org/grpc"
)

// TransactionResult represents the outcome of a transaction creation request.
type TransactionResult struct {
	Success       bool
	ErrorCode     common.TransactionErrorCode
	ErrorMessage  string
	TransactionID string
}

// TransactionRequest contains the data needed to create a new transaction.
type TransactionRequest struct {
	RecipientVSAddress  string
	Amount              uint64
	SenderPrivateKeyWIF string
}

// TransactionAdapter implements TransactionAdapterAPI using gRPC communication with the local node.
type TransactionAdapter struct {
	client  pb.AppServiceClient
	timeout time.Duration
}

// NewTransactionAdapter creates a new TransactionAdapter with the given gRPC connection.
func NewTransactionAdapter(conn grpc.ClientConnInterface) *TransactionAdapter {
	return &TransactionAdapter{
		client:  pb.NewAppServiceClient(conn),
		timeout: 30 * time.Second, // Default timeout for transaction operations
	}
}

// CreateTransaction send transaction request to local node
func (a *TransactionAdapter) CreateTransaction(ctx context.Context, req TransactionRequest) (*TransactionResult, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()

	// Create gRPC request
	grpcReq := &pb.CreateTransactionRequest{
		RecipientVsAddress:  req.RecipientVSAddress,
		Amount:              req.Amount,
		SenderPrivateKeyWif: req.SenderPrivateKeyWIF,
	}

	// Call the gRPC service
	resp, err := a.client.CreateTransaction(ctx, grpcReq)
	if err != nil {
		return nil, fmt.Errorf("gRPC call failed: %w", err)
	}

	// Map gRPC response to adapter result
	return &TransactionResult{
		Success:       resp.Success,
		ErrorCode:     mapErrorCode(resp.ErrorCode),
		ErrorMessage:  resp.ErrorMessage,
		TransactionID: resp.TransactionId,
	}, nil
}

// mapErrorCode converts gRPC error codes to adapter error codes.
func mapErrorCode(code pb.TransactionErrorCode) common.TransactionErrorCode {
	switch code {
	case pb.TransactionErrorCode_NONE:
		return common.ErrorCodeNone
	case pb.TransactionErrorCode_INVALID_PRIVATE_KEY:
		return common.ErrorCodeInvalidPrivateKey
	case pb.TransactionErrorCode_INSUFFICIENT_FUNDS:
		return common.ErrorCodeInsufficientFunds
	case pb.TransactionErrorCode_VALIDATION_FAILED:
		return common.ErrorCodeValidationFailed
	case pb.TransactionErrorCode_BROADCAST_FAILED:
		return common.ErrorCodeBroadcastFailed
	default:
		return common.ErrorCodeValidationFailed
	}
}
