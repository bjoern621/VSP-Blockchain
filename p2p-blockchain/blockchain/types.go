package blockchain

type Hash [32]byte

type InvType int

const (
	InvType_MSG_TX InvType = iota
	InvType_MSG_BLOCK
	InvType_MSG_FILTERED_BLOCK
)

type InvVector struct {
	InvType InvType
	Hash    *Hash
}

type InvMsg struct {
	Inventory []*InvVector
}
