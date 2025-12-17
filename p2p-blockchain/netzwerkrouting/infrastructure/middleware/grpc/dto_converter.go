package grpc

import (
	"fmt"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api/blockchain/dto"
)

func InvTypeDTOFromPB(t pb.InvType) (dto.InvTypeDTO, error) {
	switch t {
	case pb.InvType_MSG_TX:
		return dto.InvTypeDTO_MSG_TX, nil
	case pb.InvType_MSG_BLOCK:
		return dto.InvTypeDTO_MSG_BLOCK, nil
	case pb.InvType_MSG_FILTERED_BLOCK:
		return dto.InvTypeDTO_MSG_FILTERED_BLOCK, nil
	default:
		return 0, fmt.Errorf("unknown pb.InvType: %v", t)
	}
}

func NewBlockHeadersDTOFromPB(m *pb.BlockHeaders) (dto.BlockHeadersDTO, error) {
	if m == nil {
		return dto.BlockHeadersDTO{}, fmt.Errorf("BlockHeaders is nil")
	}
	headers := make([]dto.BlockHeaderDTO, 0, len(m.Headers))
	for i := range m.Headers {
		h, err := NewBlockHeaderDTOFromPB(m.Headers[i])
		if err != nil {
			return dto.BlockHeadersDTO{}, fmt.Errorf("headers[%d]: %w", i, err)
		}
		headers = append(headers, h)
	}
	return dto.BlockHeadersDTO{Headers: headers}, nil
}

func NewHash(bytes []byte) (dto.Hash, error) {
	if (len(bytes)) != common.HashSize {
		return dto.Hash{}, fmt.Errorf("invalid hash length")
	}
	var hash dto.Hash
	copy(hash[:], bytes)
	return hash, nil
}

func NewPublicKeyHash(bytes []byte) (dto.PublicKeyHash, error) {
	if (len(bytes)) != common.PublicKeyHashSize {
		return dto.PublicKeyHash{}, fmt.Errorf("invalid public key hash length")
	}
	var hash dto.PublicKeyHash
	copy(hash[:], bytes)
	return hash, nil
}

func NewInvVectorDTOFromPB(m *pb.InvVector) (dto.InvVectorDTO, error) {
	if m == nil {
		return dto.InvVectorDTO{}, fmt.Errorf("InvVector is nil")
	}
	t, err := InvTypeDTOFromPB(m.Type)
	if err != nil {
		return dto.InvVectorDTO{}, err
	}
	hash, err := NewHash(m.Hash)
	if err != nil {
		return dto.InvVectorDTO{}, err
	}
	return dto.InvVectorDTO{
		Type: t,
		Hash: hash,
	}, nil
}

func NewBlockLocatorDTOFromPB(m *pb.BlockLocator) (dto.BlockLocatorDTO, error) {
	if m == nil {
		return dto.BlockLocatorDTO{}, fmt.Errorf("BlockLocator is nil")
	}
	hashes := make([]dto.Hash, 0, len(m.BlockLocatorHashes))
	for i := range m.BlockLocatorHashes {
		hash, err := NewHash(m.BlockLocatorHashes[i])
		if err != nil {
			return dto.BlockLocatorDTO{}, fmt.Errorf("blockLocatorHashes[%d]: %w", i, err)
		}
		hashes = append(hashes, hash)
	}
	hash, err := NewHash(m.HashStop)
	if err != nil {
		return dto.BlockLocatorDTO{}, fmt.Errorf("hashStop: %w", err)
	}
	return dto.BlockLocatorDTO{
		BlockLocatorHashes: hashes,
		HashStop:           hash,
	}, nil
}

func NewInvMsgDTOFromPB(m *pb.InvMsg) (dto.InvMsgDTO, error) {
	if m == nil {
		return dto.InvMsgDTO{}, fmt.Errorf("InvMsg is nil")
	}
	inv := make([]dto.InvVectorDTO, 0, len(m.Inventory))
	for i := range m.Inventory {
		v, err := NewInvVectorDTOFromPB(m.Inventory[i])
		if err != nil {
			return dto.InvMsgDTO{}, fmt.Errorf("inventory[%d]: %w", i, err)
		}
		inv = append(inv, v)
	}
	return dto.InvMsgDTO{Inventory: inv}, nil
}

func NewGetDataMsgDTOFromPB(m *pb.GetDataMsg) (dto.GetDataMsgDTO, error) {
	if m == nil {
		return dto.GetDataMsgDTO{}, fmt.Errorf("GetDataMsg is nil")
	}
	inv := make([]dto.InvVectorDTO, 0, len(m.Inventory))
	for i := range m.Inventory {
		v, err := NewInvVectorDTOFromPB(m.Inventory[i])
		if err != nil {
			return dto.GetDataMsgDTO{}, fmt.Errorf("inventory[%d]: %w", i, err)
		}
		inv = append(inv, v)
	}
	return dto.GetDataMsgDTO{Inventory: inv}, nil
}

func NewBlockMsgDTOFromPB(m *pb.BlockMsg) (dto.BlockMsgDTO, error) {
	if m == nil {
		return dto.BlockMsgDTO{}, fmt.Errorf("BlockMsg is nil")
	}
	if m.Block == nil {
		return dto.BlockMsgDTO{}, fmt.Errorf("BlockMsg.Block is nil")
	}
	b, err := NewBlockDTOFromPB(m.Block)
	if err != nil {
		return dto.BlockMsgDTO{}, err
	}
	return dto.BlockMsgDTO{Block: b}, nil
}

func NewTxMsgDTOFromPB(m *pb.TxMsg) (dto.TxMsgDTO, error) {
	if m == nil {
		return dto.TxMsgDTO{}, fmt.Errorf("TxMsg is nil")
	}
	if m.Transaction == nil {
		return dto.TxMsgDTO{}, fmt.Errorf("TxMsg.Transaction is nil")
	}
	tx, err := NewTransactionDTOFromPB(m.Transaction)
	if err != nil {
		return dto.TxMsgDTO{}, err
	}
	return dto.TxMsgDTO{Transaction: tx}, nil
}

func NewSetFilterRequestDTOFromPB(m *pb.SetFilterRequest) (dto.SetFilterRequestDTO, error) {
	if m == nil {
		return dto.SetFilterRequestDTO{}, fmt.Errorf("SetFilterRequest is nil")
	}
	hashes := make([]dto.PublicKeyHash, 0, len(m.PublicKeyHashes))
	for i := range m.PublicKeyHashes {
		publicKeyHash, err := NewPublicKeyHash(m.PublicKeyHashes[i])
		if err != nil {
			return dto.SetFilterRequestDTO{}, fmt.Errorf("publicKeyHashes[%d]: %w", i, err)
		}
		hashes = append(hashes, publicKeyHash)
	}
	return dto.SetFilterRequestDTO{PublicKeyHashes: hashes}, nil
}

func NewTransactionDTOFromPB(m *pb.Transaction) (dto.TransactionDTO, error) {
	if m == nil {
		return dto.TransactionDTO{}, fmt.Errorf("transaction is nil")
	}
	inputs := make([]dto.TxInputDTO, 0, len(m.Inputs))
	for i := range m.Inputs {
		in, err := NewTxInputDTOFromPB(m.Inputs[i])
		if err != nil {
			return dto.TransactionDTO{}, fmt.Errorf("inputs[%d]: %w", i, err)
		}
		inputs = append(inputs, in)
	}
	outputs := make([]dto.TxOutputDTO, 0, len(m.Outputs))
	for i := range m.Outputs {
		out, err := NewTxOutputDTOFromPB(m.Outputs[i])
		if err != nil {
			return dto.TransactionDTO{}, fmt.Errorf("outputs[%d]: %w", i, err)
		}
		outputs = append(outputs, out)
	}
	return dto.TransactionDTO{
		Inputs:   inputs,
		Outputs:  outputs,
		LockTime: m.LockTime,
	}, nil
}

func NewTxInputDTOFromPB(m *pb.TxInput) (dto.TxInputDTO, error) {
	if m == nil {
		return dto.TxInputDTO{}, fmt.Errorf("TxInput is nil")
	}
	prevTxHash, err := NewHash(m.PrevTxHash)
	if err != nil {
		return dto.TxInputDTO{}, err
	}
	return dto.TxInputDTO{
		PrevTxHash:      prevTxHash,
		OutputIndex:     m.OutputIndex,
		SignatureScript: append([]byte(nil), m.SignatureScript...),
		Sequence:        m.Sequence,
	}, nil
}

func NewTxOutputDTOFromPB(m *pb.TxOutput) (dto.TxOutputDTO, error) {
	if m == nil {
		return dto.TxOutputDTO{}, fmt.Errorf("TxOutput is nil")
	}
	return dto.TxOutputDTO{
		Value:           m.Value,
		PublicKeyScript: append([]byte(nil), m.PublicKeyScript...),
	}, nil
}

func NewBlockHeaderDTOFromPB(m *pb.BlockHeader) (dto.BlockHeaderDTO, error) {
	if m == nil {
		return dto.BlockHeaderDTO{}, fmt.Errorf("BlockHeader is nil")
	}
	prevBlockHash, err := NewHash(m.PrevBlockHash)
	if err != nil {
		return dto.BlockHeaderDTO{}, err
	}
	merkleRoot, err := NewHash(m.MerkleRoot)
	if err != nil {
		return dto.BlockHeaderDTO{}, err
	}
	return dto.BlockHeaderDTO{
		PrevBlockHash:    prevBlockHash,
		MerkleRoot:       merkleRoot,
		Timestamp:        m.Timestamp,
		DifficultyTarget: m.DifficultyTarget,
		Nonce:            m.Nonce,
	}, nil
}

func NewBlockDTOFromPB(m *pb.Block) (dto.BlockDTO, error) {
	if m == nil {
		return dto.BlockDTO{}, fmt.Errorf("block is nil")
	}
	if m.Header == nil {
		return dto.BlockDTO{}, fmt.Errorf("block.Header is nil")
	}
	h, err := NewBlockHeaderDTOFromPB(m.Header)
	if err != nil {
		return dto.BlockDTO{}, err
	}
	txs := make([]dto.TransactionDTO, 0, len(m.Transactions))
	for i := range m.Transactions {
		tx, err := NewTransactionDTOFromPB(m.Transactions[i])
		if err != nil {
			return dto.BlockDTO{}, fmt.Errorf("transactions[%d]: %w", i, err)
		}
		txs = append(txs, tx)
	}
	return dto.BlockDTO{
		Header:       h,
		Transactions: txs,
	}, nil
}

func NewMerkleBlockMsgDTOFromPB(m *pb.MerkleBlock) (dto.MerkleBlockMsgDTO, error) {
	if m == nil {
		return dto.MerkleBlockMsgDTO{}, fmt.Errorf("MerkleBlock is nil")
	}
	if m.Header == nil {
		return dto.MerkleBlockMsgDTO{}, fmt.Errorf("MerkleBlock.Header is nil")
	}
	h, err := NewBlockHeaderDTOFromPB(m.Header)
	if err != nil {
		return dto.MerkleBlockMsgDTO{}, err
	}
	proofs := make([]dto.MerkleProofDTO, 0, len(m.Proofs))
	for i := range m.Proofs {
		p, err := NewMerkleProofDTOFromPB(m.Proofs[i])
		if err != nil {
			return dto.MerkleBlockMsgDTO{}, fmt.Errorf("proofs[%d]: %w", i, err)
		}
		proofs = append(proofs, p)
	}
	return dto.MerkleBlockMsgDTO{
		MerkleBlock: dto.MerkleBlockDTO{
			Header: h,
			Proofs: proofs,
		},
	}, nil
}

func NewMerkleProofDTOFromPB(m *pb.MerkleProof) (dto.MerkleProofDTO, error) {
	if m == nil {
		return dto.MerkleProofDTO{}, fmt.Errorf("MerkleProof is nil")
	}
	if m.Transaction == nil {
		return dto.MerkleProofDTO{}, fmt.Errorf("MerkleProof.Transaction is nil")
	}
	tx, err := NewTransactionDTOFromPB(m.Transaction)
	if err != nil {
		return dto.MerkleProofDTO{}, err
	}
	siblings := make([]dto.Hash, 0, len(m.Siblings))
	for i := range m.Siblings {
		hash, err := NewHash(m.Siblings[i])
		if err != nil {
			return dto.MerkleProofDTO{}, fmt.Errorf("siblings[%d]: %w", i, err)
		}
		siblings = append(siblings, hash)
	}
	return dto.MerkleProofDTO{
		Transaction: tx,
		Siblings:    siblings,
		Index:       m.Index,
	}, nil
}
