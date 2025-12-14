package dto

import (
	"fmt"

	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"
)

const (
	PublicKeyHashSize = 20
	PublicKeySize     = 33
	HashSize          = 32
)

type InvTypeDTO int32

const (
	InvTypeDTO_MSG_TX             InvTypeDTO = 0
	InvTypeDTO_MSG_BLOCK          InvTypeDTO = 1
	InvTypeDTO_MSG_FILTERED_BLOCK InvTypeDTO = 2
)

func InvTypeDTOFromPB(t pb.InvType) (InvTypeDTO, error) {
	switch t {
	case pb.InvType_MSG_TX:
		return InvTypeDTO_MSG_TX, nil
	case pb.InvType_MSG_BLOCK:
		return InvTypeDTO_MSG_BLOCK, nil
	case pb.InvType_MSG_FILTERED_BLOCK:
		return InvTypeDTO_MSG_FILTERED_BLOCK, nil
	default:
		return 0, fmt.Errorf("unknown pb.InvType: %v", t)
	}
}

type BlockHeadersDTO struct {
	Headers []BlockHeaderDTO
}

func NewBlockHeadersDTOFromPB(m *pb.BlockHeaders) (BlockHeadersDTO, error) {
	if m == nil {
		return BlockHeadersDTO{}, fmt.Errorf("BlockHeaders is nil")
	}
	headers := make([]BlockHeaderDTO, 0, len(m.Headers))
	for i := range m.Headers {
		h, err := NewBlockHeaderDTOFromPB(m.Headers[i])
		if err != nil {
			return BlockHeadersDTO{}, fmt.Errorf("headers[%d]: %w", i, err)
		}
		headers = append(headers, h)
	}
	return BlockHeadersDTO{Headers: headers}, nil
}

type Hash [HashSize]byte

func NewHash(bytes []byte) (Hash, error) {
	if (len(bytes)) != HashSize {
		return Hash{}, fmt.Errorf("invalid hash length")
	}
	var hash Hash
	copy(hash[:], bytes)
	return hash, nil
}

type PublicKeyHash [PublicKeyHashSize]byte

func NewPublicKeyHash(bytes []byte) (PublicKeyHash, error) {
	if (len(bytes)) != PublicKeyHashSize {
		return PublicKeyHash{}, fmt.Errorf("invalid public key hash length")
	}
	var hash PublicKeyHash
	copy(hash[:], bytes)
	return hash, nil
}

type InvVectorDTO struct {
	Type InvTypeDTO
	Hash Hash
}

func NewInvVectorDTOFromPB(m *pb.InvVector) (InvVectorDTO, error) {
	if m == nil {
		return InvVectorDTO{}, fmt.Errorf("InvVector is nil")
	}
	t, err := InvTypeDTOFromPB(m.Type)
	if err != nil {
		return InvVectorDTO{}, err
	}
	hash, err := NewHash(m.Hash)
	if err != nil {
		return InvVectorDTO{}, err
	}
	return InvVectorDTO{
		Type: t,
		Hash: hash,
	}, nil
}

type BlockLocatorDTO struct {
	BlockLocatorHashes []Hash
	HashStop           Hash
}

func NewBlockLocatorDTOFromPB(m *pb.BlockLocator) (BlockLocatorDTO, error) {
	if m == nil {
		return BlockLocatorDTO{}, fmt.Errorf("BlockLocator is nil")
	}
	hashes := make([]Hash, 0, len(m.BlockLocatorHashes))
	for i := range m.BlockLocatorHashes {
		hash, err := NewHash(m.BlockLocatorHashes[i])
		if err != nil {
			return BlockLocatorDTO{}, fmt.Errorf("blockLocatorHashes[%d]: %w", i, err)
		}
		hashes = append(hashes, hash)
	}
	hash, err := NewHash(m.HashStop)
	if err != nil {
		return BlockLocatorDTO{}, fmt.Errorf("hashStop: %w", err)
	}
	return BlockLocatorDTO{
		BlockLocatorHashes: hashes,
		HashStop:           hash,
	}, nil
}

type InvMsgDTO struct {
	Inventory []InvVectorDTO
}

func NewInvMsgDTOFromPB(m *pb.InvMsg) (InvMsgDTO, error) {
	if m == nil {
		return InvMsgDTO{}, fmt.Errorf("InvMsg is nil")
	}
	inv := make([]InvVectorDTO, 0, len(m.Inventory))
	for i := range m.Inventory {
		v, err := NewInvVectorDTOFromPB(m.Inventory[i])
		if err != nil {
			return InvMsgDTO{}, fmt.Errorf("inventory[%d]: %w", i, err)
		}
		inv = append(inv, v)
	}
	return InvMsgDTO{Inventory: inv}, nil
}

type GetDataMsgDTO struct {
	Inventory []InvVectorDTO
}

func NewGetDataMsgDTOFromPB(m *pb.GetDataMsg) (GetDataMsgDTO, error) {
	if m == nil {
		return GetDataMsgDTO{}, fmt.Errorf("GetDataMsg is nil")
	}
	inv := make([]InvVectorDTO, 0, len(m.Inventory))
	for i := range m.Inventory {
		v, err := NewInvVectorDTOFromPB(m.Inventory[i])
		if err != nil {
			return GetDataMsgDTO{}, fmt.Errorf("inventory[%d]: %w", i, err)
		}
		inv = append(inv, v)
	}
	return GetDataMsgDTO{Inventory: inv}, nil
}

type BlockMsgDTO struct {
	Block BlockDTO
}

func NewBlockMsgDTOFromPB(m *pb.BlockMsg) (BlockMsgDTO, error) {
	if m == nil {
		return BlockMsgDTO{}, fmt.Errorf("BlockMsg is nil")
	}
	if m.Block == nil {
		return BlockMsgDTO{}, fmt.Errorf("BlockMsg.Block is nil")
	}
	b, err := NewBlockDTOFromPB(m.Block)
	if err != nil {
		return BlockMsgDTO{}, err
	}
	return BlockMsgDTO{Block: b}, nil
}

type MerkleBlockMsgDTO struct {
	MerkleBlock MerkleBlockDTO
}

func NewMerkleBlockMsgDTOFromPB(m *pb.MerkleBlockMsg) (MerkleBlockMsgDTO, error) {
	if m == nil {
		return MerkleBlockMsgDTO{}, fmt.Errorf("MerkleBlockMsg is nil")
	}
	if m.MerkleBlock == nil {
		return MerkleBlockMsgDTO{}, fmt.Errorf("MerkleBlockMsg.MerkleBlock is nil")
	}
	mb, err := NewMerkleBlockDTOFromPB(m.MerkleBlock)
	if err != nil {
		return MerkleBlockMsgDTO{}, err
	}
	return MerkleBlockMsgDTO{MerkleBlock: mb}, nil
}

type TxMsgDTO struct {
	Transaction TransactionDTO
}

func NewTxMsgDTOFromPB(m *pb.TxMsg) (TxMsgDTO, error) {
	if m == nil {
		return TxMsgDTO{}, fmt.Errorf("TxMsg is nil")
	}
	if m.Transaction == nil {
		return TxMsgDTO{}, fmt.Errorf("TxMsg.Transaction is nil")
	}
	tx, err := NewTransactionDTOFromPB(m.Transaction)
	if err != nil {
		return TxMsgDTO{}, err
	}
	return TxMsgDTO{Transaction: tx}, nil
}

type SetFilterRequestDTO struct {
	PublicKeyHashes []PublicKeyHash
}

func NewSetFilterRequestDTOFromPB(m *pb.SetFilterRequest) (SetFilterRequestDTO, error) {
	if m == nil {
		return SetFilterRequestDTO{}, fmt.Errorf("SetFilterRequest is nil")
	}
	hashes := make([]PublicKeyHash, 0, len(m.PublicKeyHashes))
	for i := range m.PublicKeyHashes {
		publicKeyHash, err := NewPublicKeyHash(m.PublicKeyHashes[i])
		if err != nil {
			return SetFilterRequestDTO{}, fmt.Errorf("publicKeyHashes[%d]: %w", i, err)
		}
		hashes = append(hashes, publicKeyHash)
	}
	return SetFilterRequestDTO{PublicKeyHashes: hashes}, nil
}

type TransactionDTO struct {
	Inputs   []TxInputDTO
	Outputs  []TxOutputDTO
	LockTime uint64
}

func NewTransactionDTOFromPB(m *pb.Transaction) (TransactionDTO, error) {
	if m == nil {
		return TransactionDTO{}, fmt.Errorf("transaction is nil")
	}
	inputs := make([]TxInputDTO, 0, len(m.Inputs))
	for i := range m.Inputs {
		in, err := NewTxInputDTOFromPB(m.Inputs[i])
		if err != nil {
			return TransactionDTO{}, fmt.Errorf("inputs[%d]: %w", i, err)
		}
		inputs = append(inputs, in)
	}
	outputs := make([]TxOutputDTO, 0, len(m.Outputs))
	for i := range m.Outputs {
		out, err := NewTxOutputDTOFromPB(m.Outputs[i])
		if err != nil {
			return TransactionDTO{}, fmt.Errorf("outputs[%d]: %w", i, err)
		}
		outputs = append(outputs, out)
	}
	return TransactionDTO{
		Inputs:   inputs,
		Outputs:  outputs,
		LockTime: m.LockTime,
	}, nil
}

type TxInputDTO struct {
	PrevTxHash      Hash
	OutputIndex     uint32
	SignatureScript []byte // Wie groß ist?
	Sequence        uint32
}

func NewTxInputDTOFromPB(m *pb.TxInput) (TxInputDTO, error) {
	if m == nil {
		return TxInputDTO{}, fmt.Errorf("TxInput is nil")
	}
	prevTxHash, err := NewHash(m.PrevTxHash)
	if err != nil {
		return TxInputDTO{}, err
	}
	return TxInputDTO{
		PrevTxHash:      prevTxHash,
		OutputIndex:     m.OutputIndex,
		SignatureScript: append([]byte(nil), m.SignatureScript...),
		Sequence:        m.Sequence,
	}, nil
}

type TxOutputDTO struct {
	Value           uint64
	PublicKeyScript []byte // Wie groß ist?
}

func NewTxOutputDTOFromPB(m *pb.TxOutput) (TxOutputDTO, error) {
	if m == nil {
		return TxOutputDTO{}, fmt.Errorf("TxOutput is nil")
	}
	return TxOutputDTO{
		Value:           m.Value,
		PublicKeyScript: append([]byte(nil), m.PublicKeyScript...),
	}, nil
}

type BlockHeaderDTO struct {
	PrevBlockHash    Hash
	MerkleRoot       Hash
	Timestamp        int64
	DifficultyTarget uint32
	Nonce            uint32
}

func NewBlockHeaderDTOFromPB(m *pb.BlockHeader) (BlockHeaderDTO, error) {
	if m == nil {
		return BlockHeaderDTO{}, fmt.Errorf("BlockHeader is nil")
	}
	prevBlockHash, err := NewHash(m.PrevBlockHash)
	if err != nil {
		return BlockHeaderDTO{}, err
	}
	merkleRoot, err := NewHash(m.MerkleRoot)
	if err != nil {
		return BlockHeaderDTO{}, err
	}
	return BlockHeaderDTO{
		PrevBlockHash:    prevBlockHash,
		MerkleRoot:       merkleRoot,
		Timestamp:        m.Timestamp,
		DifficultyTarget: m.DifficultyTarget,
		Nonce:            m.Nonce,
	}, nil
}

type BlockDTO struct {
	Header       BlockHeaderDTO
	Transactions []TransactionDTO
}

func NewBlockDTOFromPB(m *pb.Block) (BlockDTO, error) {
	if m == nil {
		return BlockDTO{}, fmt.Errorf("block is nil")
	}
	if m.Header == nil {
		return BlockDTO{}, fmt.Errorf("block.Header is nil")
	}
	h, err := NewBlockHeaderDTOFromPB(m.Header)
	if err != nil {
		return BlockDTO{}, err
	}
	txs := make([]TransactionDTO, 0, len(m.Transactions))
	for i := range m.Transactions {
		tx, err := NewTransactionDTOFromPB(m.Transactions[i])
		if err != nil {
			return BlockDTO{}, fmt.Errorf("transactions[%d]: %w", i, err)
		}
		txs = append(txs, tx)
	}
	return BlockDTO{
		Header:       h,
		Transactions: txs,
	}, nil
}

type MerkleBlockDTO struct {
	Header BlockHeaderDTO
	Proofs []MerkleProofDTO
}

func NewMerkleBlockDTOFromPB(m *pb.MerkleBlock) (MerkleBlockDTO, error) {
	if m == nil {
		return MerkleBlockDTO{}, fmt.Errorf("MerkleBlock is nil")
	}
	if m.Header == nil {
		return MerkleBlockDTO{}, fmt.Errorf("MerkleBlock.Header is nil")
	}
	h, err := NewBlockHeaderDTOFromPB(m.Header)
	if err != nil {
		return MerkleBlockDTO{}, err
	}
	proofs := make([]MerkleProofDTO, 0, len(m.Proofs))
	for i := range m.Proofs {
		p, err := NewMerkleProofDTOFromPB(m.Proofs[i])
		if err != nil {
			return MerkleBlockDTO{}, fmt.Errorf("proofs[%d]: %w", i, err)
		}
		proofs = append(proofs, p)
	}
	return MerkleBlockDTO{
		Header: h,
		Proofs: proofs,
	}, nil
}

type MerkleProofDTO struct {
	Transaction TransactionDTO
	Siblings    []Hash
	Index       uint32
}

func NewMerkleProofDTOFromPB(m *pb.MerkleProof) (MerkleProofDTO, error) {
	if m == nil {
		return MerkleProofDTO{}, fmt.Errorf("MerkleProof is nil")
	}
	if m.Transaction == nil {
		return MerkleProofDTO{}, fmt.Errorf("MerkleProof.Transaction is nil")
	}
	tx, err := NewTransactionDTOFromPB(m.Transaction)
	if err != nil {
		return MerkleProofDTO{}, err
	}
	siblings := make([]Hash, 0, len(m.Siblings))
	for i := range m.Siblings {
		hash, err := NewHash(m.Siblings[i])
		if err != nil {
			return MerkleProofDTO{}, fmt.Errorf("siblings[%d]: %w", i, err)
		}
		siblings = append(siblings, hash)
	}
	return MerkleProofDTO{
		Transaction: tx,
		Siblings:    siblings,
		Index:       m.Index,
	}, nil
}
