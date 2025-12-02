package grpc

import (
	"context"
	"net/netip"
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/handshake"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"

	"bjoernblessin.de/go-utils/util/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

// versionInfoFromProto converts protobuf VersionInfo to domain VersionInfo.
func versionInfoFromProto(req *pb.VersionInfo) handshake.VersionInfo {
	var endpoint netip.AddrPort
	if req.ListeningEndpoint != nil {
		if ip, ok := netip.AddrFromSlice(req.ListeningEndpoint.IpAddress); ok {
			endpoint = netip.AddrPortFrom(ip, uint16(req.ListeningEndpoint.ListeningPort))
		}
	}

	services := make([]peer.ServiceType, len(req.SupportedServices))
	for i, pbService := range req.SupportedServices {
		services[i] = peer.ServiceType(pbService)
	}

	return handshake.VersionInfo{
		Version:           req.GetVersion(),
		SupportedServices: services,
		ListeningEndpoint: endpoint,
	}
}

func versionInfoToProto(info handshake.VersionInfo) *pb.VersionInfo {
	pbInfo := &pb.VersionInfo{
		Version: info.Version,
		ListeningEndpoint: &pb.Endpoint{
			IpAddress:     info.ListeningEndpoint.Addr().AsSlice(),
			ListeningPort: uint32(info.ListeningEndpoint.Port()),
		},
	}
	for _, service := range info.SupportedServices {
		pbInfo.SupportedServices = append(pbInfo.SupportedServices, pb.ServiceType(service))
	}
	return pbInfo
}

func (s *Server) Version(ctx context.Context, req *pb.VersionInfo) (*emptypb.Empty, error) {
	peerAddrPort := getPeerAddr(ctx)
	peerID := s.peerRegistry.GetOrCreatePeerID(peerAddrPort)
	info := versionInfoFromProto(req)

	s.connectionHandler.HandleVersion(peerID, info)
	return &emptypb.Empty{}, nil
}

func (s *Server) Verack(ctx context.Context, req *pb.VersionInfo) (*emptypb.Empty, error) {
	peerAddrPort := getPeerAddr(ctx)
	peerID := s.peerRegistry.GetOrCreatePeerID(peerAddrPort)
	info := versionInfoFromProto(req)

	s.connectionHandler.HandleVerack(peerID, info)
	return &emptypb.Empty{}, nil
}

func (s *Server) Ack(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	peerAddrPort := getPeerAddr(ctx)
	peerID := s.peerRegistry.GetOrCreatePeerID(peerAddrPort)

	s.connectionHandler.HandleAck(peerID)
	return &emptypb.Empty{}, nil
}

func (c *Client) SendVersion(peerID peer.PeerID, info handshake.VersionInfo, addrPort netip.AddrPort) {
	conn, err := grpc.NewClient(addrPort.String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Errorf("failed to connect to %s: %v", addrPort.String(), err)
		return
	}

	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			logger.Errorf("failed to close gRPC connection: %v", err)
		}
	}(conn)

	c.peerRegistry.AddPeer(peerID, addrPort)

	client := pb.NewConnectionEstablishmentClient(conn)

	pbInfo := versionInfoToProto(info)

	_, err = client.Version(context.Background(), pbInfo)
	if err != nil {
		logger.Errorf("failed to send Version to %s: %v", addrPort.String(), err)
	}
}

func (c *Client) SendVerack(peerID peer.PeerID, info handshake.VersionInfo) {

}

func (c *Client) SendAck(peerID peer.PeerID) {

}
