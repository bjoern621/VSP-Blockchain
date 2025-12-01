package grpc

import (
	"context"
	"net/netip"
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"

	"bjoernblessin.de/go-utils/util/logger"
	grpcPeer "google.golang.org/grpc/peer"
	"google.golang.org/protobuf/types/known/emptypb"
)

// getPeerAddr extracts the remote peer address from the gRPC context.
func getPeerAddr(ctx context.Context) netip.AddrPort {
	p, ok := grpcPeer.FromContext(ctx)
	if !ok {
		logger.Errorf("could not get peer from context")
	}

	addrStr := p.Addr.String()
	addrPort := netip.MustParseAddrPort(addrStr)

	return addrPort
}

// toVersionInfo converts protobuf VersionInfo to domain VersionInfo.
func toVersionInfo(req *pb.VersionInfo) peer.VersionInfo {
	var endpoint netip.AddrPort
	if req.ListeningEndpoint != nil {
		if ip, ok := netip.AddrFromSlice(req.ListeningEndpoint.IpAddress); ok {
			endpoint = netip.AddrPortFrom(ip, uint16(req.ListeningEndpoint.ListeningPort))
		}
	}

	services := make([]string, len(req.SupportedServices))
	for i, svc := range req.SupportedServices {
		services[i] = svc.String()
	}

	return peer.VersionInfo{
		Version:           req.Version,
		SupportedServices: services,
		ListeningEndpoint: endpoint,
	}
}

func (s *Server) Version(ctx context.Context, req *pb.VersionInfo) (*emptypb.Empty, error) {
	peerAddrPort := getPeerAddr(ctx)
	peerID := s.peerRegistry.GetOrCreatePeerID(peerAddrPort)
	info := toVersionInfo(req)

	s.connectionHandler.HandleVersion(peerID, info)
	return &emptypb.Empty{}, nil
}

func (s *Server) Verack(ctx context.Context, req *pb.VersionInfo) (*emptypb.Empty, error) {
	peerAddrPort := getPeerAddr(ctx)
	peerID := s.peerRegistry.GetOrCreatePeerID(peerAddrPort)
	info := toVersionInfo(req)

	s.connectionHandler.HandleVerack(peerID, info)
	return &emptypb.Empty{}, nil
}

func (s *Server) Ack(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	peerAddrPort := getPeerAddr(ctx)
	peerID := s.peerRegistry.GetOrCreatePeerID(peerAddrPort)

	s.connectionHandler.HandleAck(peerID)
	return &emptypb.Empty{}, nil
}
