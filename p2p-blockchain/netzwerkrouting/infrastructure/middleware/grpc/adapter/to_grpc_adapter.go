package adapter

import (
	"fmt"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/inv"
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"
)

func ToGrpcGetDataMsg(inventory []*inv.InvVector) (*pb.GetDataMsg, error) {
	if inventory == nil {
		return nil, fmt.Errorf("inventory must not be nil")
	}
	pbInv := toGrpcInvVector(inventory)
	return &pb.GetDataMsg{Inventory: pbInv}, nil
}

func ToGrpcGetInvMsg(inventory []*inv.InvVector) (*pb.InvMsg, error) {
	if inventory == nil {
		return nil, fmt.Errorf("inventory must not be nil")
	}
	pbInv := toGrpcInvVector(inventory)
	return &pb.InvMsg{Inventory: pbInv}, nil
}

func ToGrpcBlockLocator(locator block.BlockLocator) (*pb.BlockLocator, error) {
	if locator.BlockLocatorHashes == nil {
		return nil, fmt.Errorf("block locator hashes must not be nil")
	}

	hashes := make([][]byte, len(locator.BlockLocatorHashes))
	for i, hash := range locator.BlockLocatorHashes {
		hashes[i] = hash[:]
	}

	return &pb.BlockLocator{
		BlockLocatorHashes: hashes,
		HashStop:           locator.StopHash[:],
	}, nil
}

func toGrpcInvVector(inventory []*inv.InvVector) []*pb.InvVector {
	pbInv := make([]*pb.InvVector, len(inventory))
	for i, v := range inventory {
		pbInv[i] = &pb.InvVector{
			Type: pb.InvType(v.InvType),
			Hash: v.Hash[:],
		}
	}
	return pbInv
}

func ToGrpcHeadersMsg(headers []*block.BlockHeader) (*pb.BlockHeaders, error) {
	if headers == nil {
		return nil, fmt.Errorf("headers must not be nil")
	}

	pbHeaders := make([]*pb.BlockHeader, len(headers))
	for i, h := range headers {
		if h == nil {
			return nil, fmt.Errorf("headers[%d] must not be nil", i)
		}
		pbHeaders[i] = &pb.BlockHeader{
			PrevBlockHash:    h.PreviousBlockHash[:],
			MerkleRoot:       h.MerkleRoot[:],
			Timestamp:        h.Timestamp,
			DifficultyTarget: uint32(h.DifficultyTarget),
			Nonce:            h.Nonce,
		}
	}

	return &pb.BlockHeaders{Headers: pbHeaders}, nil
}
