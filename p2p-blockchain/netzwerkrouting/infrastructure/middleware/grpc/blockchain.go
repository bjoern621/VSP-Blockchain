package grpc

import (
	"context"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/inv"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/infrastructure/middleware/grpc/adapter"

	"bjoernblessin.de/go-utils/util/logger"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) Inv(ctx context.Context, msg *pb.InvMsg) (*emptypb.Empty, error) {
	peerId := s.GetPeerId(ctx)

	invMsgDTO, err := adapter.ToInvVectorsFromInvMsg(msg)
	if err != nil {
		return nil, err
	}

	go s.NotifyInv(invMsgDTO, peerId)

	return &emptypb.Empty{}, nil
}

func (s *Server) GetData(ctx context.Context, msg *pb.GetDataMsg) (*emptypb.Empty, error) {
	peerId := s.GetPeerId(ctx)

	getDataMsgDTO, err := adapter.ToInvVectorsFromGetDataMsg(msg)
	if err != nil {
		return nil, err
	}

	go s.NotifyGetData(getDataMsgDTO, peerId)

	return &emptypb.Empty{}, nil
}

func (s *Server) Block(ctx context.Context, msg *pb.BlockMsg) (*emptypb.Empty, error) {
	peerId := s.GetPeerId(ctx)

	blockMsgDTO, err := adapter.ToBlockFromBlockMsg(msg)
	if err != nil {
		return nil, err
	}

	go s.NotifyBlock(blockMsgDTO, peerId)

	return &emptypb.Empty{}, nil
}

func (s *Server) MerkleBlock(ctx context.Context, msg *pb.MerkleBlockMsg) (*emptypb.Empty, error) {
	peerId := s.GetPeerId(ctx)

	merkleBlockMsgDTO, err := adapter.ToMerkleBlockFromMerkleBlockMsg(msg)
	if err != nil {
		return nil, err
	}

	go s.NotifyMerkleBlock(merkleBlockMsgDTO, peerId)

	return &emptypb.Empty{}, nil
}

func (s *Server) Tx(ctx context.Context, msg *pb.TxMsg) (*emptypb.Empty, error) {
	peerId := s.GetPeerId(ctx)

	txMsgDTO, err := adapter.ToTxFromTxMsg(msg)
	if err != nil {
		return nil, err
	}

	go s.NotifyTx(txMsgDTO, peerId)

	return &emptypb.Empty{}, nil
}

func (s *Server) GetHeaders(ctx context.Context, locator *pb.BlockLocator) (*emptypb.Empty, error) {
	peerId := s.GetPeerId(ctx)

	blockLocator, err := adapter.ToBlockLocator(locator)
	if err != nil {
		return nil, err
	}

	go s.NotifyGetHeaders(blockLocator, peerId)

	return &emptypb.Empty{}, nil
}

func (s *Server) Headers(ctx context.Context, pbHeaders *pb.BlockHeaders) (*emptypb.Empty, error) {
	peerId := s.GetPeerId(ctx)

	headers, err := adapter.ToHeadersFromHeadersMsg(pbHeaders)
	if err != nil {
		return nil, err
	}

	go s.NotifyHeaders(headers, peerId)

	return &emptypb.Empty{}, nil
}

func (s *Server) SetFilter(ctx context.Context, request *pb.SetFilterRequest) (*emptypb.Empty, error) {
	peerId := s.GetPeerId(ctx)

	filterRequest, err := adapter.ToSetFilterRequestFromSetFilterRequest(request)
	if err != nil {
		return nil, err
	}

	go s.NotifySetFilterRequest(filterRequest, peerId)

	return &emptypb.Empty{}, nil
}

func (s *Server) Mempool(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	peerId := s.GetPeerId(ctx)
	go s.NotifyMempool(peerId)

	return &emptypb.Empty{}, nil
}

func (s *Server) NotifyInv(inventory []*inv.InvVector, peerID common.PeerId) {
	for observer := range s.observers.Iter() {
		observer.Inv(inventory, peerID)
	}
}

func (s *Server) NotifyGetData(inventory []*inv.InvVector, peerID common.PeerId) {
	for observer := range s.observers.Iter() {
		observer.GetData(inventory, peerID)
	}
}

func (s *Server) NotifyBlock(block block.Block, peerID common.PeerId) {
	for observer := range s.observers.Iter() {
		observer.Block(block, peerID)
	}
}

func (s *Server) NotifyMerkleBlock(merkleBlock block.MerkleBlock, peerID common.PeerId) {
	for observer := range s.observers.Iter() {
		observer.MerkleBlock(merkleBlock, peerID)
	}
}

func (s *Server) NotifyTx(tx transaction.Transaction, peerID common.PeerId) {
	for observer := range s.observers.Iter() {
		observer.Tx(tx, peerID)
	}
}

func (s *Server) NotifyGetHeaders(locator block.BlockLocator, peerID common.PeerId) {
	for observer := range s.observers.Iter() {
		observer.GetHeaders(locator, peerID)
	}
}

func (s *Server) NotifyHeaders(headers []*block.BlockHeader, peerID common.PeerId) {
	for observer := range s.observers.Iter() {
		observer.Headers(headers, peerID)
	}
}

func (s *Server) NotifySetFilterRequest(setFilterRequest block.SetFilterRequest, peerID common.PeerId) {
	for observer := range s.observers.Iter() {
		observer.SetFilter(setFilterRequest, peerID)
	}
}

func (s *Server) NotifyMempool(peerID common.PeerId) {
	for observer := range s.observers.Iter() {
		observer.Mempool(peerID)
	}
}

// SendGetData sends a getdata message to the given peer
func (c *Client) SendGetData(inv []*inv.InvVector, peerId common.PeerId) {
	pbMsg, err := adapter.ToGrpcGetDataMsg(inv)
	if err != nil {
		logger.Warnf("failed to create GetDataMessage from DTO: %v", err)
		return
	}

	SendHelper(c, peerId, "GetData", pb.NewBlockchainServiceClient, func(client pb.BlockchainServiceClient) error {
		_, err := client.GetData(context.Background(), pbMsg)
		return err
	})
}

// SendInv sends an inv message to the given peer
func (c *Client) SendInv(inv []*inv.InvVector, peerId common.PeerId) {
	pbInvMsg, err := adapter.ToGrpcGetInvMsg(inv)
	if err != nil {
		logger.Warnf("failed to create InvMessage from DTO: %v", err)
		return
	}

	SendHelper(c, peerId, "Inv", pb.NewBlockchainServiceClient, func(client pb.BlockchainServiceClient) error {
		_, err := client.Inv(context.Background(), pbInvMsg)
		return err
	})
}

// SendGetHeaders sends a GetHeaders message to the given peer
func (c *Client) SendGetHeaders(locator block.BlockLocator, peerId common.PeerId) {
	pbLocator, err := adapter.ToGrpcBlockLocator(locator)
	if err != nil {
		logger.Warnf("failed to create BlockLocator from DTO: %v", err)
		return
	}

	SendHelper(c, peerId, "GetHeaders", pb.NewBlockchainServiceClient, func(client pb.BlockchainServiceClient) error {
		_, err := client.GetHeaders(context.Background(), pbLocator)
		return err
	})
}

// SendHeaders sends a Headers message to the given peer
func (c *Client) SendHeaders(headers []*block.BlockHeader, peerId common.PeerId) {
	pbHeaders, err := adapter.ToGrpcHeadersMsg(headers)
	if err != nil {
		logger.Warnf("failed to create HeadersMsg from DTO: %v", err)
		return
	}

	go SendHelper(c, peerId, "Headers", pb.NewBlockchainServiceClient, func(client pb.BlockchainServiceClient) error {
		_, err := client.Headers(context.Background(), pbHeaders)
		return err
	})
}

func (c *Client) SendBlock(block block.Block, peerId common.PeerId) {
	pbBlockMsg, err := adapter.ToGrpcBlockMsg(&block)
	if err != nil {
		logger.Warnf("failed to create BlockMsg from DTO: %v", err)
		return
	}

	SendHelper(c, peerId, "Block", pb.NewBlockchainServiceClient, func(client pb.BlockchainServiceClient) error {
		_, err := client.Block(context.Background(), pbBlockMsg)
		return err
	})
}

func (c *Client) SendTx(transaction transaction.Transaction, peerId common.PeerId) {
	pbTxMsg, err := adapter.ToGrpcTxMsg(&transaction)
	if err != nil {
		logger.Warnf("failed to create TxMsg from DTO: %v", err)
		return
	}

	SendHelper(c, peerId, "Tx", pb.NewBlockchainServiceClient, func(client pb.BlockchainServiceClient) error {
		_, err := client.Tx(context.Background(), pbTxMsg)
		return err
	})
}
