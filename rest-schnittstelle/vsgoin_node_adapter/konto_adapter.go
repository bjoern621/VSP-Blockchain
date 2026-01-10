// Package adapter implements the konto adapter pattern for asset queries.
// It encapsulates all interactions with the local V$Goin SPV Node via gRPC.
package vsgoin_node_adapter

import (
	"context"
	"fmt"
	"s3b/vsp-blockchain/rest-api/internal/common"
	"s3b/vsp-blockchain/rest-api/internal/pb"

	"google.golang.org/grpc"
)

// KontoAdapterAPI provides the interface for querying assets from the local V$Goin Node.
type KontoAdapterAPI interface {
	// GetAssets queries the assets for a given V$Address via the local node.
	GetAssets(vsAddress string) (*common.AssetsResult, error)
}

// KontoAdapter implements KontoAdapterAPI using gRPC communication with the local node.
type KontoAdapter struct {
	client pb.AppServiceClient
}

// NewKontoAdapter creates a new KontoAdapter with the given gRPC connection.
func NewKontoAdapter(conn grpc.ClientConnInterface) *KontoAdapter {
	return &KontoAdapter{
		client: pb.NewAppServiceClient(conn),
	}
}

// GetAssets queries the assets for a given V$Address via the local node.
func (a *KontoAdapter) GetAssets(vsAddress string) (*common.AssetsResult, error) {

	grpcReq := &pb.GetAssetsRequest{
		VsAddress: vsAddress,
	}

	resp, err := a.client.GetAssets(context.Background(), grpcReq)
	if err != nil {
		return nil, fmt.Errorf("gRPC call failed: %w", err)
	}

	// Convert protobuf assets to domain assets
	assets := make([]common.Asset, 0, len(resp.Assets))
	for _, pbAsset := range resp.Assets {
		assets = append(assets, common.Asset{
			Value: pbAsset.Value,
		})
	}

	return &common.AssetsResult{
		Success:      resp.Success,
		ErrorMessage: resp.ErrorMessage,
		Assets:       assets,
	}, nil
}
