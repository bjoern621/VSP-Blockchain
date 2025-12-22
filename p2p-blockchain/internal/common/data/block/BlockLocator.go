package block

import "s3b/vsp-blockchain/p2p-blockchain/internal/common"

type BlockLocator struct {
	BlockLocatorHashes []common.Hash
	StopHash           common.Hash
}
