package grpc

import (
	"context"
	"fmt"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain"
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"

	"bjoernblessin.de/go-utils/util/assert"
	"google.golang.org/protobuf/types/known/emptypb"
)

func bytesToHash(bytes []byte) (*blockchain.Hash, error) {
	var hash blockchain.Hash
	if len(bytes) != 32 {
		return &hash, fmt.Errorf("invalid hash length: %d", len(bytes))
	}

	copy(hash[:], bytes)
	return &hash, nil
}

func fromProtoInvVector(vector *pb.InvVector) *blockchain.InvVector {
	hash, err := bytesToHash(vector.Hash)
	assert.IsNil(err, "failed to convert hash from proto")

	return &blockchain.InvVector{
		InvType: blockchain.InvType(vector.Type),
		Hash:    hash,
	}
}

func fromProtoInvVectors(inventoryVector []*pb.InvVector) []*blockchain.InvVector {
	invVectors := make([]*blockchain.InvVector, len(inventoryVector))
	for _, pbInvVector := range inventoryVector {
		invVectors = append(invVectors, fromProtoInvVector(pbInvVector))
	}

	return invVectors
}

func (s *Server) Inv(ctx context.Context, msg *pb.InvMsg) (*emptypb.Empty, error) {
	invVectors := fromProtoInvVectors(msg.Inventory)
	go s.NotifyInv(blockchain.InvMsg{
		Inventory: invVectors,
	})

	return &emptypb.Empty{}, nil
}

func (s *Server) GetData(ctx context.Context, msg *pb.GetDataMsg) (*emptypb.Empty, error) {
	invVectors := fromProtoInvVectors(msg.Inventory)
	go s.NotifyGetData(blockchain.GetDataMsg{
		Inventory: invVectors,
	})

	return &emptypb.Empty{}, nil
}

func (s *Server) Block(ctx context.Context, msg *pb.BlockMsg) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Server) MerkleBlock(ctx context.Context, msg *pb.MerkleBlockMsg) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Server) Tx(ctx context.Context, msg *pb.TxMsg) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Server) GetHeaders(ctx context.Context, locator *pb.BlockLocator) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Server) Headers(ctx context.Context, headers *pb.BlockHeaders) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Server) SetFilter(ctx context.Context, request *pb.SetFilterRequest) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Server) Mempool(ctx context.Context, empty *emptypb.Empty) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Server) NotifyInv(invMsg blockchain.InvMsg) {
	for observer := range s.observers {
		observer.Inv(invMsg)
	}
}

func (s *Server) NotifyGetData(getDataMsg blockchain.GetDataMsg) {
	for observer := range s.observers {
		observer.GetData(getDataMsg)
	}
}
