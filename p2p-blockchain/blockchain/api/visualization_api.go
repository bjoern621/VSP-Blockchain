package api

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core"
)

// VisualizationAPI provides access to blockchain visualization functionality.
type VisualizationAPI interface {
	// GetVisualizationDot returns a Graphviz DOT format string representing the blockchain structure.
	// If includeDetails is true, nodes will include height and accumulated work information.
	GetVisualizationDot(includeDetails bool) string
}

// VisualizationAPIImpl implements VisualizationAPI using the BlockStore.
type VisualizationAPIImpl struct {
	blockStore core.BlockStoreAPI
}

// NewVisualizationAPI creates a new VisualizationAPI with the given BlockStore.
func NewVisualizationAPI(blockStore core.BlockStoreAPI) *VisualizationAPIImpl {
	return &VisualizationAPIImpl{
		blockStore: blockStore,
	}
}

// GetVisualizationDot implements VisualizationAPI.GetVisualizationDot.
func (v *VisualizationAPIImpl) GetVisualizationDot(includeDetails bool) string {
	return v.blockStore.GetVisualizationDot(includeDetails)
}
