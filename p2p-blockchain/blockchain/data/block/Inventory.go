package block

import (
	"fmt"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api/blockchain/dto"
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
