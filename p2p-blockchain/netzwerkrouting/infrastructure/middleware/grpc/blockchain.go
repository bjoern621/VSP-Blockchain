package grpc

import (
	"context"
	"net/netip"
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api/blockchain/dto"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"

	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) GetPeerId(ctx context.Context) peer.PeerID {
	inboundAddr := GetPeerAddr(ctx)
	peerID := s.networkInfoRegistry.GetOrRegisterPeer(inboundAddr, netip.AddrPort{})
	s.networkInfoRegistry.AddInboundAddress(peerID, inboundAddr)

	return peerID
}

func (s *Server) Inv(ctx context.Context, msg *pb.InvMsg) (*emptypb.Empty, error) {
	peerId := s.GetPeerId(ctx)

	invMsgDTO, err := dto.NewInvMsgDTOFromPB(msg)
	if err != nil {
		return nil, err
	}

	go s.NotifyInv(invMsgDTO, peerId)

	return &emptypb.Empty{}, nil
}

func (s *Server) GetData(ctx context.Context, msg *pb.GetDataMsg) (*emptypb.Empty, error) {
	peerId := s.GetPeerId(ctx)

	getDataMsgDTO, err := dto.NewGetDataMsgDTOFromPB(msg)
	if err != nil {
		return nil, err
	}

	go s.NotifyGetData(getDataMsgDTO, peerId)

	return &emptypb.Empty{}, nil
}

func (s *Server) Block(ctx context.Context, msg *pb.BlockMsg) (*emptypb.Empty, error) {
	peerId := s.GetPeerId(ctx)

	blockMsgDTO, err := dto.NewBlockMsgDTOFromPB(msg)
	if err != nil {
		return nil, err
	}

	go s.NotifyBlock(blockMsgDTO, peerId)

	return &emptypb.Empty{}, nil
}

func (s *Server) MerkleBlock(ctx context.Context, msg *pb.MerkleBlockMsg) (*emptypb.Empty, error) {
	peerId := s.GetPeerId(ctx)

	merkleBlockMsgDTO, err := dto.NewMerkleBlockMsgDTOFromPB(msg)
	if err != nil {
		return nil, err
	}

	go s.NotifyMerkleBlock(merkleBlockMsgDTO, peerId)

	return &emptypb.Empty{}, nil
}

func (s *Server) Tx(ctx context.Context, msg *pb.TxMsg) (*emptypb.Empty, error) {
	peerId := s.GetPeerId(ctx)

	txMsgDTO, err := dto.NewTxMsgDTOFromPB(msg)
	if err != nil {
		return nil, err
	}

	go s.NotifyTx(txMsgDTO, peerId)

	return &emptypb.Empty{}, nil
}

func (s *Server) GetHeaders(ctx context.Context, locator *pb.BlockLocator) (*emptypb.Empty, error) {
	peerId := s.GetPeerId(ctx)

	blockLocator, err := dto.NewBlockLocatorDTOFromPB(locator)
	if err != nil {
		return nil, err
	}

	go s.NotifyGetHeaders(blockLocator, peerId)

	return &emptypb.Empty{}, nil
}

func (s *Server) Headers(ctx context.Context, pbHeaders *pb.BlockHeaders) (*emptypb.Empty, error) {
	peerId := s.GetPeerId(ctx)

	headers, err := dto.NewBlockHeadersDTOFromPB(pbHeaders)
	if err != nil {
		return nil, err
	}

	go s.NotifyHeaders(headers, peerId)

	return &emptypb.Empty{}, nil
}

func (s *Server) SetFilter(ctx context.Context, request *pb.SetFilterRequest) (*emptypb.Empty, error) {
	peerId := s.GetPeerId(ctx)

	filterRequest, err := dto.NewSetFilterRequestDTOFromPB(request)
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

func (s *Server) NotifyInv(invMsg dto.InvMsgDTO, peerID peer.PeerID) {
	for observer := range s.observers {
		observer.Inv(invMsg, peerID)
	}
}

func (s *Server) NotifyGetData(getDataMsg dto.GetDataMsgDTO, peerID peer.PeerID) {
	for observer := range s.observers {
		observer.GetData(getDataMsg, peerID)
	}
}

func (s *Server) NotifyBlock(blockMsg dto.BlockMsgDTO, peerID peer.PeerID) {
	for observer := range s.observers {
		observer.Block(blockMsg, peerID)
	}
}

func (s *Server) NotifyMerkleBlock(merkleBlockMsg dto.MerkleBlockMsgDTO, peerID peer.PeerID) {
	for observer := range s.observers {
		observer.MerkleBlock(merkleBlockMsg, peerID)
	}
}

func (s *Server) NotifyTx(txMsg dto.TxMsgDTO, peerID peer.PeerID) {
	for observer := range s.observers {
		observer.Tx(txMsg, peerID)
	}
}

func (s *Server) NotifyGetHeaders(locator dto.BlockLocatorDTO, peerID peer.PeerID) {
	for observer := range s.observers {
		observer.GetHeaders(locator, peerID)
	}
}

func (s *Server) NotifyHeaders(headers dto.BlockHeadersDTO, peerID peer.PeerID) {
	for observer := range s.observers {
		observer.Headers(headers, peerID)
	}
}

func (s *Server) NotifySetFilterRequest(setFilterRequest dto.SetFilterRequestDTO, peerID peer.PeerID) {
	for observer := range s.observers {
		observer.SetFilter(setFilterRequest, peerID)
	}
}

func (s *Server) NotifyMempool(peerID peer.PeerID) {
	for observer := range s.observers {
		observer.Mempool(peerID)
	}
}
