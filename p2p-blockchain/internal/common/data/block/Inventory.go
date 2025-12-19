package block

type InvType int

const (
	InvTypeMsgTx            InvType = 0
	InvTypeMsgBlock         InvType = 1
	InvTypeMsgFilteredBlock InvType = 2
)

type InvVector struct {
	InvType InvType
	Hash    Hash
}
