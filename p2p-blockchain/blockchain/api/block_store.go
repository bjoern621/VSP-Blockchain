package api

import "s3b/vsp-blockchain/p2p-blockchain/internal/common"

type BlockStore interface {
	IsKnownBlockHash(hash common.Hash) bool
}
