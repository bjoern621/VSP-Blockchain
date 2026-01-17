package api

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
)

type BlockStoreAPI interface {
	core.BlockStoreAPI
}

// BlockStoreVisualizationAPI defines the interface for accessing block data needed for visualization.
type BlockStoreVisualizationAPI interface {
	GetAllBlocksWithMetadata() []block.BlockWithMetadata
}
