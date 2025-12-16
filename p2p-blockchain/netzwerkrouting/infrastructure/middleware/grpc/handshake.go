package grpc

import (
	"context"
	"net/netip"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
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
func versionInfoFromProto(info *pb.VersionInfo) (handshake.VersionInfo, netip.AddrPort) {
	var endpoint netip.AddrPort
	if info.ListeningEndpoint != nil {
		if ip, ok := netip.AddrFromSlice(info.ListeningEndpoint.IpAddress); ok {
			endpoint = netip.AddrPortFrom(ip, uint16(info.ListeningEndpoint.ListeningPort))
		}
	}

	services := make([]peer.ServiceType, len(info.SupportedServices))
	for i, pbService := range info.SupportedServices {
		services[i] = peer.ServiceType(pbService)
	}

	return handshake.VersionInfo{
		Version:           info.GetVersion(),
		SupportedServices: services,
	}, endpoint
}

func versionInfoToProto(info handshake.VersionInfo, addrPort netip.AddrPort) *pb.VersionInfo {
	pbInfo := &pb.VersionInfo{
		Version: info.Version,
		ListeningEndpoint: &pb.Endpoint{
			IpAddress:     addrPort.Addr().AsSlice(),
			ListeningPort: uint32(addrPort.Port()),
		},
	}
	for _, service := range info.SupportedServices {
		pbInfo.SupportedServices = append(pbInfo.SupportedServices, pb.ServiceType(service))
	}
	return pbInfo
}

func createGRPCClient(remoteAddrPort netip.AddrPort) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(remoteAddrPort.String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Warnf("failed to connect to %s: %v", remoteAddrPort.String(), err)
		return nil, err
	}
	return conn, nil
}

func (s *Server) Version(ctx context.Context, req *pb.VersionInfo) (*emptypb.Empty, error) {
	inboundAddr := getPeerAddr(ctx)
	info, addrPort := versionInfoFromProto(req)

	peerID := s.networkInfoRegistry.GetOrRegisterPeer(inboundAddr, addrPort)
	s.networkInfoRegistry.AddInboundAddress(peerID, inboundAddr)
	s.networkInfoRegistry.SetListeningEndpoint(peerID, addrPort)

	s.handshakeMsgHandler.HandleVersion(peerID, info)
	return &emptypb.Empty{}, nil
}

func (s *Server) Verack(ctx context.Context, req *pb.VersionInfo) (*emptypb.Empty, error) {
	inboundAddr := getPeerAddr(ctx)
	info, addrPort := versionInfoFromProto(req)

	peerID := s.networkInfoRegistry.GetOrRegisterPeer(inboundAddr, addrPort)
	s.networkInfoRegistry.AddInboundAddress(peerID, inboundAddr)
	s.networkInfoRegistry.SetListeningEndpoint(peerID, addrPort)

	s.handshakeMsgHandler.HandleVerack(peerID, info)
	return &emptypb.Empty{}, nil
}

func (s *Server) Ack(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	peerId := s.GetPeerId(ctx)

	s.handshakeMsgHandler.HandleAck(peerId)
	return &emptypb.Empty{}, nil
}

func (c *Client) SendVersion(peerID common.PeerId, localInfo handshake.VersionInfo) {
	remoteAddrPort, ok := c.networkInfoRegistry.GetListeningEndpoint(peerID)
	if !ok {
		logger.Warnf("failed to send Version: no listening endpoint for peer %s", peerID)
		return
	}

	conn, err := createGRPCClient(remoteAddrPort)
	if err != nil {
		logger.Warnf("failed to create gRPC client for %s: %v", remoteAddrPort.String(), err)
		return
	}

	localAddrPort := netip.AddrPortFrom(common.P2PListeningIpAddr(), common.P2PPort())

	c.networkInfoRegistry.SetConnection(peerID, conn)

	client := pb.NewConnectionEstablishmentClient(conn)

	pbInfo := versionInfoToProto(localInfo, localAddrPort)

	_, err = client.Version(context.Background(), pbInfo)
	if err != nil {
		logger.Warnf("failed to send Version to %s: %v", remoteAddrPort.String(), err)
	}
}

func (c *Client) SendVerack(peerID common.PeerId, localInfo handshake.VersionInfo) {
	remoteAddrPort, ok := c.networkInfoRegistry.GetListeningEndpoint(peerID)
	if !ok {
		logger.Warnf("failed to send Verack: no listening endpoint for peer %s", peerID)
		return
	}

	conn, err := createGRPCClient(remoteAddrPort)
	if err != nil {
		logger.Warnf("failed to create gRPC client for %s: %v", remoteAddrPort.String(), err)
		return
	}

	localAddrPort := netip.AddrPortFrom(common.P2PListeningIpAddr(), common.P2PPort())

	c.networkInfoRegistry.SetConnection(peerID, conn)

	client := pb.NewConnectionEstablishmentClient(conn)

	pbInfo := versionInfoToProto(localInfo, localAddrPort)

	_, err = client.Verack(context.Background(), pbInfo)
	if err != nil {
		logger.Warnf("failed to send Verack to %s: %v", peerID, err)
	}
}

func (c *Client) SendAck(peerID common.PeerId) {
	conn, ok := c.networkInfoRegistry.GetConnection(peerID)
	if !ok {
		logger.Warnf("failed to send Ack: no connection for peer %s", peerID)
		return
	}

	client := pb.NewConnectionEstablishmentClient(conn)

	_, err := client.Ack(context.Background(), &emptypb.Empty{})
	if err != nil {
		logger.Warnf("failed to send Ack to peer %s: %v", peerID, err)
	}
}
