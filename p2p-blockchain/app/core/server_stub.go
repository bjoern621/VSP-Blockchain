package core

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"

	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/infrastructure/middleware/grpc/peerregistry"

	"bjoernblessin.de/go-utils/util/logger"
	"google.golang.org/grpc"
)

// Server represents the app gRPC server for external local systems.
type Server struct {
	pb.UnimplementedAppServiceServer
	grpcServer   *grpc.Server
	listener     net.Listener
	handshakeAPI api.HandshakeAPI
	peerRegistry *peerregistry.PeerRegistry
}

// NewServer creates a new external API server.
func NewServer(handshakeAPI api.HandshakeAPI, peerRegistry *peerregistry.PeerRegistry) *Server {
	return &Server{
		handshakeAPI: handshakeAPI,
		peerRegistry: peerRegistry,
	}
}

// Start starts the external API gRPC server on the given port in a goroutine.
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

	port := uint16(req.Port)
	if req.Port > 65535 {
		return &pb.ConnectToResponse{
			Success:      false,
			ErrorMessage: "port must be between 0 and 65535",
		}, nil
	}

	addrPort := netip.AddrPortFrom(ip, port)

	s.handshakeAPI.InitiateHandshake(addrPort)

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
func (s *Server) GetPeerRegistry(ctx context.Context, req *pb.GetPeerRegistryRequest) (*pb.GetPeerRegistryResponse, error) {
	entries := s.peerRegistry.GetAllEntries()

	response := &pb.GetPeerRegistryResponse{
		Entries: make([]*pb.PeerRegistryEntry, 0, len(entries)),
	}

	for peerID, addrPort := range entries {
		response.Entries = append(response.Entries, &pb.PeerRegistryEntry{
			PeerId:    string(peerID),
			IpAddress: addrPort.Addr().AsSlice(),
			Port:      uint32(addrPort.Port()),
		})
	}

	return response, nil
}
