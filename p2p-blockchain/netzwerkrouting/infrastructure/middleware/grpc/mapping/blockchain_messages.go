package mapping

import (
	"fmt"
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api/blockchain/dto"
)

func InvTypePBFromDTO(t dto.InvTypeDTO) (pb.InvType, error) {
	switch t {
	case dto.InvTypeDTO_MSG_TX:
		return pb.InvType_MSG_TX, nil
	case dto.InvTypeDTO_MSG_BLOCK:
		return pb.InvType_MSG_BLOCK, nil
	case dto.InvTypeDTO_MSG_FILTERED_BLOCK:
		return pb.InvType_MSG_FILTERED_BLOCK, nil
	default:
		return 0, fmt.Errorf("unknown dto.InvTypeDTO: %v", t)
	}
}

func NewInvVectorPBFromDTO(v dto.InvVectorDTO) (*pb.InvVector, error) {
	t, err := InvTypePBFromDTO(v.Type)
	if err != nil {
		return nil, err
	}
	return &pb.InvVector{
		Type: t,
		Hash: append([]byte(nil), v.Hash[:]...),
	}, nil
}

func NewGetDataMessageFromDTO(m dto.GetDataMsgDTO) (*pb.GetDataMsg, error) {
	inv := make([]*pb.InvVector, 0, len(m.Inventory))
	for i := range m.Inventory {
		v, err := NewInvVectorPBFromDTO(m.Inventory[i])
		if err != nil {
			return nil, fmt.Errorf("inventory[%d]: %w", i, err)
		}
		inv = append(inv, v)
	}
	return &pb.GetDataMsg{Inventory: inv}, nil
}

func NewInvMessageFromDTO(m dto.InvMsgDTO) (*pb.InvMsg, error) {
	inv := make([]*pb.InvVector, 0, len(m.Inventory))
	for i := range m.Inventory {
		v, err := NewInvVectorPBFromDTO(m.Inventory[i])
		if err != nil {
			return nil, fmt.Errorf("inventory[%d]: %w", i, err)
		}

		inv = append(inv, v)
	}

	return &pb.InvMsg{Inventory: inv}, nil
}
