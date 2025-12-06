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

func fromProtoInvMsg(msg *pb.InvMsg) blockchain.InvMsg {
	invVector := make([]*blockchain.InvVector, len(msg.Inventory))
	for _, pbInvVector := range msg.Inventory {
		hash, err := bytesToHash(pbInvVector.Hash)
		assert.IsNil(err, "failed to convert hash from proto")

		invVector = append(invVector, &blockchain.InvVector{
			InvType: blockchain.InvType(pbInvVector.Type),
			Hash:    hash,
		})
	}

	invMsg := blockchain.InvMsg{
		Inventory: invVector,
	}
	return invMsg
}

func (s *Server) Inv(ctx context.Context, msg *pb.InvMsg) (*emptypb.Empty, error) {
	invMsg := fromProtoInvMsg(msg)
	go s.NotifyInv(invMsg)

	return &emptypb.Empty{}, nil
}

func (s *Server) NotifyInv(invMsg blockchain.InvMsg) {
	for observer := range s.observers {
		observer.Inv(invMsg)
	}
}

func (s *Server) GetData(ctx context.Context, msg *pb.GetDataMsg) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
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
