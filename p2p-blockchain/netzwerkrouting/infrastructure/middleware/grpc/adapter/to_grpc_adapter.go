package adapter

import (
	"fmt"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"
)

func ToGrpcGetDataMsg(inventory []*block.InvVector) (*pb.GetDataMsg, error) {
	if inventory == nil {
		return nil, fmt.Errorf("inventory must not be nil")
	}
	pbInv := toGrpcInvVector(inventory)
	return &pb.GetDataMsg{Inventory: pbInv}, nil
}

func ToGrpcGetInvMsg(inventory []*block.InvVector) (*pb.InvMsg, error) {
	if inventory == nil {
		return nil, fmt.Errorf("inventory must not be nil")
	}
	pbInv := toGrpcInvVector(inventory)
	return &pb.InvMsg{Inventory: pbInv}, nil
}

func toGrpcInvVector(inventory []*block.InvVector) []*pb.InvVector {
	pbInv := make([]*pb.InvVector, len(inventory))
	for i, v := range inventory {
		pbInv[i] = &pb.InvVector{
			Type: pb.InvType(v.InvType),
			Hash: v.Hash[:],
		}
	}
	return pbInv
}
