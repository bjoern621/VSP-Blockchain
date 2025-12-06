package blockchain

type Hash [32]byte

type InvType int

const (
	InvType_MSG_TX InvType = iota
	InvType_MSG_BLOCK
	InvType_MSG_FILTERED_BLOCK
)

type InvMsg struct {
	Inventory []*InvVector
}

type GetDataMsg struct {
	Inventory []*InvVector
}

type BlockMsg struct {
	Block *Block
}

type MerkleBlockMsg struct {
	MerkleBlock *MerkleBlock
}

type TxMsg struct {
	Transaction *Transaction
}

type InvVector struct {
	InvType InvType
	Hash    *Hash
}

type BlockHeader struct {
	Hash             *Hash
	MerkleRoot       *Hash
	Timestamp        int64
	DifficultyTarget uint32
	Nonce            uint32
}

type TxInput struct {
	PreviousTxHash  *Hash
	PreviousIndex   uint32
	SignatureScript []byte
	Sequence        uint32
}

type TxOutput struct {
	Value           int64
	PublicKeyScript []byte
}

type Transaction struct {
	Inputs   []*TxInput
	Outputs  []*TxOutput
	LockTime int64
}

type Block struct {
	Header       *BlockHeader
	Transactions []*Transaction
}

type MerkleProof struct {
	Transaction *Transaction
	Siblings    []*Hash
	Index       uint32
}

type MerkleBlock struct {
	BlockHeader *BlockHeader
	Proofs      []*MerkleProof
}

type BlockLocator struct {
	BlockLocatorHashes []*Hash
	StopHash           *Hash
}
