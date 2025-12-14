package dto

import "s3b/vsp-blockchain/p2p-blockchain/internal/common"

type InvTypeDTO int32

const (
	InvTypeDTO_MSG_TX             InvTypeDTO = 0
	InvTypeDTO_MSG_BLOCK          InvTypeDTO = 1
	InvTypeDTO_MSG_FILTERED_BLOCK InvTypeDTO = 2
)

type BlockHeadersDTO struct {
	Headers []BlockHeaderDTO
}

type Hash [common.HashSize]byte

type PublicKeyHash [common.PublicKeyHashSize]byte

type InvVectorDTO struct {
	Type InvTypeDTO
	Hash Hash
}

type BlockLocatorDTO struct {
	BlockLocatorHashes []Hash
	HashStop           Hash
}

type InvMsgDTO struct {
	Inventory []InvVectorDTO
}

type GetDataMsgDTO struct {
	Inventory []InvVectorDTO
}

type BlockMsgDTO struct {
	Block BlockDTO
}

type MerkleBlockMsgDTO struct {
	MerkleBlock MerkleBlockDTO
}

type TxMsgDTO struct {
	Transaction TransactionDTO
}

type SetFilterRequestDTO struct {
	PublicKeyHashes []PublicKeyHash
}

type TransactionDTO struct {
	Inputs   []TxInputDTO
	Outputs  []TxOutputDTO
	LockTime uint64
}

type TxInputDTO struct {
	PrevTxHash      Hash
	OutputIndex     uint32
	SignatureScript []byte // Wie groß ist?
	Sequence        uint32
}

type TxOutputDTO struct {
	Value           uint64
	PublicKeyScript []byte // Wie groß ist?
}

type BlockHeaderDTO struct {
	PrevBlockHash    Hash
	MerkleRoot       Hash
	Timestamp        int64
	DifficultyTarget uint32
	Nonce            uint32
}

type BlockDTO struct {
	Header       BlockHeaderDTO
	Transactions []TransactionDTO
}

type MerkleBlockDTO struct {
	Header BlockHeaderDTO
	Proofs []MerkleProofDTO
}

type MerkleProofDTO struct {
	Transaction TransactionDTO
	Siblings    []Hash
	Index       uint32
}
