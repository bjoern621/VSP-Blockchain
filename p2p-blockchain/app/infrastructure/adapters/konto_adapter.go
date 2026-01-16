package adapters

import (
	appapi "s3b/vsp-blockchain/p2p-blockchain/app/api"
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"
)

// KontoHandlerAdapter handles konto queries from gRPC requests.
type KontoHandlerAdapter struct {
	kontoAPI appapi.KontoAPI
}

// NewKontoAdapter creates a new KontoHandlerAdapter with the given konto API.
func NewKontoAdapter(api appapi.KontoAPI) *KontoHandlerAdapter {
	return &KontoHandlerAdapter{
		kontoAPI: api,
	}
}

// GetAssets handles the GetAssets RPC call from external local systems.
func (k *KontoHandlerAdapter) GetAssets(req *pb.GetAssetsRequest) *pb.GetAssetsResponse {
	if req.VsAddress == "" {
		return &pb.GetAssetsResponse{
			Success:      false,
			ErrorMessage: "V$Address is required",
		}
	}

	result := k.kontoAPI.GetAssets(req.VsAddress)

	if !result.Success {
		return &pb.GetAssetsResponse{
			Success:      false,
			ErrorMessage: result.ErrorMessage,
		}
	}

	// Convert assets to protobuf format
	pbAssets := make([]*pb.Asset, 0, len(result.Assets))
	for _, asset := range result.Assets {
		pbAssets = append(pbAssets, &pb.Asset{
			Value: asset.Value,
		})
	}

	return &pb.GetAssetsResponse{
		Success: true,
		Assets:  pbAssets,
	}
}
