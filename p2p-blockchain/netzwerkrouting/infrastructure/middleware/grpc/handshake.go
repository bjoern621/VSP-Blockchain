package grpc

import (
	"context"
	"net/netip"
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core"

	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/types/known/emptypb"
)

// getPeerAddr extracts the remote peer address from the gRPC context.
func getPeerAddr(ctx context.Context) string {
	if p, ok := peer.FromContext(ctx); ok {
		return p.Addr.String()
	}
	return "unknown"
}

// toVersionInfo converts protobuf VersionInfo to domain VersionInfo.
func toVersionInfo(req *pb.VersionInfo) core.VersionInfo {
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

	return core.VersionInfo{
		Version:           req.Version,
		SupportedServices: services,
		ListeningEndpoint: endpoint,
	}
}

func (s *Server) Version(ctx context.Context, req *pb.VersionInfo) (*emptypb.Empty, error) {
	peerAddr := getPeerAddr(ctx)
	info := toVersionInfo(req)

	if err := s.connectionHandler.HandleVersion(ctx, peerAddr, info); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) Verack(ctx context.Context, req *pb.VersionInfo) (*emptypb.Empty, error) {
	peerAddr := getPeerAddr(ctx)
	info := toVersionInfo(req)

	if err := s.connectionHandler.HandleVerack(ctx, peerAddr, info); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) Ack(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	peerAddr := getPeerAddr(ctx)

	if err := s.connectionHandler.HandleAck(ctx, peerAddr); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
