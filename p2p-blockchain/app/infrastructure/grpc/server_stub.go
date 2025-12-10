package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"

	"s3b/vsp-blockchain/p2p-blockchain/app/core"
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"

	"bjoernblessin.de/go-utils/util/logger"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
)

// Server represents the app gRPC server for external local systems.
type Server struct {
	pb.UnimplementedAppServiceServer
	grpcServer  *grpc.Server
	listener    net.Listener
	connService *core.ConnectionEstablishmentService
	regService  *core.InternalViewService
}

// NewServer creates a new external API server.
func NewServer(connService *core.ConnectionEstablishmentService, regService *core.InternalViewService) *Server {
	return &Server{
		connService: connService,
		regService:  regService,
	}
}

// Start starts the external API gRPC server on the given port in a goroutine.
func (s *Server) Start(port uint16) error {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	s.listener = listener

	s.grpcServer = grpc.NewServer()
	pb.RegisterAppServiceServer(s.grpcServer, s)

	go func() {
		if err := s.grpcServer.Serve(listener); err != nil {
			logger.Warnf("App gRPC server stopped with error: %v", err)
		}
	}()

	return nil
}

// Stop gracefully stops the app server.
func (s *Server) Stop() {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
}

// ConnectTo handles the ConnectTo RPC call from external local systems.
func (s *Server) ConnectTo(ctx context.Context, req *pb.ConnectToRequest) (*pb.ConnectToResponse, error) {
	ip, ok := netip.AddrFromSlice(req.IpAddress)
	if !ok {
		return &pb.ConnectToResponse{
			Success:      false,
			ErrorMessage: "invalid IP address format",
		}, nil
	}

	if req.Port > 65535 {
		return &pb.ConnectToResponse{
			Success:      false,
			ErrorMessage: "port must be between 0 and 65535",
		}, nil
	}

	port := uint16(req.Port)

	err := s.connService.ConnectTo(ip, port)
	if err != nil {
		return &pb.ConnectToResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to initiate handshake: %v", err),
		}, nil
	}

	return &pb.ConnectToResponse{
		Success:      true,
		ErrorMessage: "",
	}, nil
}

// ListeningEndpoint returns the server's listening endpoint as netip.AddrPort.
// If the server is not started, it returns an error.
func (s *Server) ListeningEndpoint() (netip.AddrPort, error) {
	if s.listener == nil {
		return netip.AddrPort{}, errors.New("server not started")
	}
	addr := s.listener.Addr().(*net.TCPAddr)
	return netip.AddrPortFrom(netip.MustParseAddr(addr.IP.String()), uint16(addr.Port)), nil
}

// GetPeerRegistry returns the current peer registry for debugging purposes.
func (s *Server) GetInternalPeerInfo(ctx context.Context, req *pb.GetInternalPeerInfoRequest) (*pb.GetInternalPeerInfoResponse, error) {
	peers := s.regService.GetInternalPeerInfo()

	response := &pb.GetInternalPeerInfoResponse{
		Entries: make([]*pb.PeerRegistryEntry, 0, len(peers)),
	}

	for _, p := range peers {
		// Convert infrastructure data to structpb.Struct
		var infraStruct *structpb.Struct
		if p.PeerInfrastructureData != nil {
			// Type assert to map[string]any which is what the infrastructure layer returns
			if infraMap, ok := p.PeerInfrastructureData.(map[string]any); ok {
				var err error
				infraStruct, err = structpb.NewStruct(infraMap)
				if err != nil {
					logger.Warnf("failed to create structpb from infra data: %v", err)
					infraStruct, _ = structpb.NewStruct(nil)
				}
			}
		}

		if infraStruct == nil {
			infraStruct, _ = structpb.NewStruct(nil)
		}

		entry := &pb.PeerRegistryEntry{
			PeerId:             string(p.PeerID),
			InfrastructureData: infraStruct,
			Version:            p.Version,
			ConnectionState:    p.ConnectionState,
			Direction:          p.Direction,
			SupportedServices:  p.SupportedServices,
		}

		response.Entries = append(response.Entries, entry)
	}

	return response, nil
}
