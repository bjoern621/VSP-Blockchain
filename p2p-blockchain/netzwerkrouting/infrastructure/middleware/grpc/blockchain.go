package grpc

import (
	"context"
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"

	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) Inv(ctx context.Context, msg *pb.InvMsg) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
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
