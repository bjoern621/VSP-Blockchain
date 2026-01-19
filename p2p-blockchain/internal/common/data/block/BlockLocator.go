package block

import "s3b/vsp-blockchain/p2p-blockchain/internal/common"

type BlockLocator struct {
	// BlockLocatorHashes are ordered from newest to oldest
	BlockLocatorHashes []common.Hash
	// StopHash defines the last (highest) hash to be returned (inclusive) (zero hash means no limit)
	StopHash common.Hash
}
