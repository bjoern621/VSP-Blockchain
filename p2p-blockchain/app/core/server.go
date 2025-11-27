package core

import (
	"context"
	"fmt"
	"net"
	"net/netip"

	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"
	netzwerkroutingCore "s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core"

	"google.golang.org/grpc"
)

// Server represents the app gRPC server for external local systems.
type Server struct {
	pb.UnimplementedAppServiceServer
	grpcServer *grpc.Server
	listener   net.Listener
}

// NewServer creates a new external API server.
func NewServer() *Server {
	return &Server{}
}

// Start starts the external API gRPC server on the given port.
func (s *Server) Start(port uint16) error {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	s.listener = listener

	s.grpcServer = grpc.NewServer()
	pb.RegisterAppServiceServer(s.grpcServer, s)

	go func() {
		if err := s.grpcServer.Serve(listener); err != nil {
			// Log error but don't crash - server may have been stopped intentionally
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

	port := uint16(req.Port)
	if req.Port > 65535 {
		return &pb.ConnectToResponse{
			Success:      false,
			ErrorMessage: "port must be between 0 and 65535",
		}, nil
	}

	err := netzwerkroutingCore.ConnectTo(ctx, ip, port)
	if err != nil {
		return &pb.ConnectToResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	return &pb.ConnectToResponse{
		Success:      true,
		ErrorMessage: "",
	}, nil
}
