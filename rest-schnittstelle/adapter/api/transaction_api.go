package api

import (
	"context"
	"s3b/vsp-blockchain/rest-api/adapter/core"
)

// TransactionAdapterAPI provides the interface for interacting with the local V$Goin Node.
type TransactionAdapterAPI interface {
	// CreateTransaction creates and broadcasts a new transaction via the local node.
	CreateTransaction(ctx context.Context, req core.TransactionRequest) (*core.TransactionResult, error)
}
