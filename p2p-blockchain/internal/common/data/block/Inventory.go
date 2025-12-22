package block

import "s3b/vsp-blockchain/p2p-blockchain/internal/common"

type InvType int

const (
	InvTypeMsgTx InvType = iota
	InvTypeMsgBlock
	InvTypeMsgFilteredBlock
)

type InvVector struct {
	InvType InvType
	Hash    common.Hash
}
