package grpc

import (
	"context"
	"net/netip"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/inv"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/infrastructure/middleware/grpc/adapter"

	"bjoernblessin.de/go-utils/util/logger"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) GetPeerId(ctx context.Context) common.PeerId {
	inboundAddr := getPeerAddr(ctx)
	peerID := s.networkInfoRegistry.GetOrRegisterPeer(inboundAddr, netip.AddrPort{})
	s.networkInfoRegistry.AddInboundAddress(peerID, inboundAddr)

	return peerID
}

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

func (c *Client) SendGetData(inv []*inv.InvVector, peerId common.PeerId) {
	conn, ok := c.networkInfoRegistry.GetConnection(peerId)
	if !ok {
		logger.Warnf("failed to send GetDataMsg: no connection for peer %s", peerId)
		return
	}

	client := pb.NewBlockchainServiceClient(conn)
	pbMsg, err := adapter.ToGrpcGetDataMsg(inv)
	if err != nil {
		logger.Warnf("failed to create GetDataMessage from DTO: %v", err)
		return
	}
	go func() {
		_, err := client.GetData(context.Background(), pbMsg)
		if err != nil {
			logger.Warnf("failed to send GetDataMsg to peer %s: %v", peerId, err)
		}
	}()
}

func (c *Client) SendInv(inv []*inv.InvVector, peerId common.PeerId) {
	conn, ok := c.networkInfoRegistry.GetConnection(peerId)
	if !ok {
		logger.Warnf("failed to send InvMsg: no connection for peer %s", peerId)
		return
	}

	client := pb.NewBlockchainServiceClient(conn)
	pbInvMsg, err := adapter.ToGrpcGetInvMsg(inv)
	if err != nil {
		logger.Warnf("failed to create InvMessage from DTO: %v", err)
		return
	}
	go func() {
		_, err := client.Inv(context.Background(), pbInvMsg)
		if err != nil {
			logger.Warnf("failed to send InvMsg to peer %s: %v", peerId, err)
		}
	}()
}
