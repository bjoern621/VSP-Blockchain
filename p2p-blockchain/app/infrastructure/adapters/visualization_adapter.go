package adapters

import (
	blockapi "s3b/vsp-blockchain/p2p-blockchain/blockchain/api"
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"
)

// VisualizationHandlerAdapter handles blockchain visualization requests from gRPC.
type VisualizationHandlerAdapter struct {
	visualizationAPI blockapi.VisualizationAPI
}

// NewVisualizationAdapter creates a new VisualizationHandlerAdapter with the given visualization API.
func NewVisualizationAdapter(api blockapi.VisualizationAPI) *VisualizationHandlerAdapter {
	return &VisualizationHandlerAdapter{
		visualizationAPI: api,
	}
}

// GetBlockchainVisualization handles the GetBlockchainVisualization RPC call.
func (v *VisualizationHandlerAdapter) GetBlockchainVisualization(req *pb.GetBlockchainVisualizationRequest) *pb.GetBlockchainVisualizationResponse {
	includeDetails := false
	if req != nil {
		includeDetails = req.IncludeDetails
	}

	dotContent := v.visualizationAPI.GetVisualizationDot(includeDetails)

	return &pb.GetBlockchainVisualizationResponse{
		DotContent: dotContent,
	}
}
