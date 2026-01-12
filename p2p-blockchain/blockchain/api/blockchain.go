package api

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
)

// BlockchainAPI defines the operations that the blockchain module exposes to other modules.
type BlockchainAPI interface {
	// AddSelfMinedBlock notifies the blockchain about a newly mined block by the local miner.
	AddSelfMinedBlock(selfMinedBlock block.Block)
}
