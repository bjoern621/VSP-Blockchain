package api

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core"
)

// VisualizationAPI provides access to blockchain visualization functionality.
type VisualizationAPI interface {
	// GetVisualizationURL returns a URL to GraphvizOnline that displays the blockchain structure.
	// If includeDetails is true, nodes will include height and accumulated work information.
	GetVisualizationURL(includeDetails bool) string
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

// GetVisualizationURL implements VisualizationAPI.GetVisualizationURL.
func (v *VisualizationAPIImpl) GetVisualizationURL(includeDetails bool) string {
	return v.blockStore.GetVisualizationURL(includeDetails)
}
