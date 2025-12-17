package block

import (
	"fmt"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/transaction"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api/blockchain/dto"

	"bjoernblessin.de/go-utils/util/assert"
)

type InvType int

const (
	InvTypeMsgTx            InvType = 0
	InvTypeMsgBlock         InvType = 1
	InvTypeMsgFilteredBlock InvType = 2
)

func invTypeDTOFromPB(t dto.InvTypeDTO) (InvType, error) {
	switch t {
	case dto.InvTypeDTO_MSG_TX:
		return InvTypeMsgTx, nil
	case dto.InvTypeDTO_MSG_BLOCK:
		return InvTypeMsgBlock, nil
	case dto.InvTypeDTO_MSG_FILTERED_BLOCK:
		return InvTypeMsgFilteredBlock, nil
	default:
		return 0, fmt.Errorf("unknown dto.InvType: %v", t)
	}
}

func invTypeToDto(invType InvType) (dto.InvTypeDTO, error) {
	switch invType {
	case InvTypeMsgTx:
		return dto.InvTypeDTO_MSG_TX, nil
	case InvTypeMsgBlock:
		return dto.InvTypeDTO_MSG_BLOCK, nil
	case InvTypeMsgFilteredBlock:
		return dto.InvTypeDTO_MSG_FILTERED_BLOCK, nil
	default:
		return 0, fmt.Errorf("unknown InvType: %v", invType)
	}
}

type InvVector struct {
	InvType InvType
	Hash    Hash
}

func newInvTypeFromDTO(t dto.InvTypeDTO) InvType {
	out, err := invTypeDTOFromPB(t)
	if err != nil {
		panic(err)
	}
	return out
}

func newInvVectorFromDTO(v dto.InvVectorDTO) InvVector {
	return InvVector{
		InvType: newInvTypeFromDTO(v.Type),
		Hash:    NewHashFromDTO(v.Hash),
	}
}

func newInvVectors(dtoVector []dto.InvVectorDTO) []InvVector {
	inv := make([]InvVector, 0, len(dtoVector))
	for i := range dtoVector {
		inv = append(inv, newInvVectorFromDTO(dtoVector[i]))
	}
	return inv
}

func InvVectorsFromInvMsgDTO(invMsgDto dto.InvMsgDTO) []InvVector {
	return newInvVectors(invMsgDto.Inventory)
}

func InvVectorFromGetDataDTO(invMsgDto dto.GetDataMsgDTO) []InvVector {
	return newInvVectors(invMsgDto.Inventory)
}

func (i *InvVector) ToDtoInvVector() dto.InvVectorDTO {
	out, err := invTypeToDto(i.InvType)
	assert.IsNil(err)

	var hash dto.Hash

	copy(hash[:], i.Hash[:])

	return dto.InvVectorDTO{
		Type: out,
		Hash: hash,
	}
}

func ToDtoInvVectors(invVectors []InvVector) []dto.InvVectorDTO {
	out := make([]dto.InvVectorDTO, 0, len(invVectors))
	for i := range invVectors {
		out = append(out, invVectors[i].ToDtoInvVector())
	}
	return out
}

func FromTxToDtoInvVector(tx transaction.Transaction) dto.InvVectorDTO {
	txId := tx.Hash()
	var hash dto.Hash

	copy(hash[:], txId[:])

	return dto.InvVectorDTO{
		Type: dto.InvTypeDTO_MSG_TX,
		Hash: hash,
	}
}
