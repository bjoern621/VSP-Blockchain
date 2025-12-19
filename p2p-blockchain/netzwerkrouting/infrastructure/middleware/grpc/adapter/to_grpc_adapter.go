package adapter

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"
)

func ToGrpcGetDataMsg(inventory []*block.InvVector) (*pb.GetDataMsg, error) {
	return nil, nil
}

func ToGrpcGetInvMsg(inventory []*block.InvVector) (*pb.InvMsg, error) {
	return nil, nil
}
