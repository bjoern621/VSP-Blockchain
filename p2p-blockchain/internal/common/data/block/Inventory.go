package block

import "s3b/vsp-blockchain/p2p-blockchain/internal/common"

type InvType int

const (
	InvTypeMsgTx            InvType = 0
	InvTypeMsgBlock         InvType = 1
	InvTypeMsgFilteredBlock InvType = 2
)

type InvVector struct {
	InvType InvType
	Hash    common.Hash
}
