package block

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/constants"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/transaction"
)

type Hash [constants.HashSize]byte

type InvType int

type InvMsg struct {
	Inventory []InvVector
}

type GetDataMsg struct {
	Inventory []InvVector
}

type BlockMsg struct {
	Block Block
}

type MerkleBlockMsg struct {
	MerkleBlock MerkleBlock
}

type TxMsg struct {
	Transaction transaction.Transaction
}

type InvVector struct {
	InvType InvType
	Hash    Hash
}

type BlockHeader struct {
	Hash             Hash
	MerkleRoot       Hash
	Timestamp        int64
	DifficultyTarget uint32
	Nonce            uint32
}

type Block struct {
	Header       BlockHeader
	Transactions []transaction.Transaction
}

type MerkleProof struct {
	Transaction transaction.Transaction
	Siblings    []Hash
	Index       uint32
}

type MerkleBlock struct {
	BlockHeader BlockHeader
	Proofs      []MerkleProof
}

type BlockLocator struct {
	BlockLocatorHashes []Hash
	StopHash           Hash
}

type SetFilterRequest struct {
	PublicKeyHashes []Hash
}
