package adapters

import (
	"s3b/vsp-blockchain/p2p-blockchain/app/core"
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"
)

// VisualizationHandlerAdapter handles blockchain visualization requests from gRPC.
type VisualizationHandlerAdapter struct {
	visualizationService *core.VisualizationService
}

// NewVisualizationAdapter creates a new VisualizationHandlerAdapter with the given visualization service.
func NewVisualizationAdapter(service *core.VisualizationService) *VisualizationHandlerAdapter {
	return &VisualizationHandlerAdapter{
		visualizationService: service,
	}
}

// GetBlockchainVisualization handles the GetBlockchainVisualization RPC call.
func (v *VisualizationHandlerAdapter) GetBlockchainVisualization(req *pb.GetBlockchainVisualizationRequest) *pb.GetBlockchainVisualizationResponse {
	includeDetails := false
	if req != nil {
		includeDetails = req.IncludeDetails
	}

	visualizationURL := v.visualizationService.GetVisualizationURL(includeDetails)

	return &pb.GetBlockchainVisualizationResponse{
		VisualizationUrl: visualizationURL,
	}
}
