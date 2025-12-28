package api

import (
	"context"
	"s3b/vsp-blockchain/rest-api/internal/common"
)

// KontoAdapterAPI provides the interface for querying assets from the local V$Goin Node.
type KontoAdapterAPI interface {
	// GetAssets queries the assets for a given V$Address via the local node.
	GetAssets(ctx context.Context, vsAddress string) (*common.AssetsResult, error)
}
