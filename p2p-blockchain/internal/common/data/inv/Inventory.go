package inv

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

func (iv InvVector) String() string {
	typeStr := ""
	switch iv.InvType {
	case InvTypeMsgTx:
		typeStr = "tx"
	case InvTypeMsgBlock:
		typeStr = "block"
	case InvTypeMsgFilteredBlock:
		typeStr = "filteredblock"
	}
	return iv.Hash.String() + " (" + typeStr + ")"
}
