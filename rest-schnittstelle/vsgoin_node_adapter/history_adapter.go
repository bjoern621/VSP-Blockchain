// Package vsgoin_node_adapter implements the history adapter pattern for transaction history queries.
// It encapsulates all interactions with the local V$Goin SPV Node via gRPC.
package vsgoin_node_adapter

import (
	"context"
	"fmt"
	"s3b/vsp-blockchain/rest-api/internal/common"
	"s3b/vsp-blockchain/rest-api/internal/pb"

	"google.golang.org/grpc"
)

// HistoryAdapterAPI provides the interface for querying transaction history from the local V$Goin Node.
type HistoryAdapterAPI interface {
	// GetHistory queries the transaction history for a given V$Address via the local node.
	GetHistory(vsAddress string) (*common.HistoryResult, error)
}

// HistoryAdapter implements HistoryAdapterAPI using gRPC communication with the local node.
type HistoryAdapter struct {
	client pb.AppServiceClient
}

// NewHistoryAdapter creates a new HistoryAdapter with the given gRPC connection.
func NewHistoryAdapter(conn grpc.ClientConnInterface) *HistoryAdapter {
	return &HistoryAdapter{
		client: pb.NewAppServiceClient(conn),
	}
}

// GetHistory queries the transaction history for a given V$Address via the local node.
func (a *HistoryAdapter) GetHistory(vsAddress string) (*common.HistoryResult, error) {
	grpcReq := &pb.GetHistoryRequest{
		VsAddress: vsAddress,
	}

	resp, err := a.client.GetHistory(context.Background(), grpcReq)
	if err != nil {
		return nil, fmt.Errorf("gRPC call failed: %w", err)
	}

	return &common.HistoryResult{
		Success:      resp.Success,
		ErrorMessage: resp.ErrorMessage,
		Transactions: resp.Transactions,
	}, nil
}
