package grpc

import (
	"errors"
	"fmt"
	"net"
	"net/netip"

	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/handshake"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/infrastructure/middleware/grpc/peerregistry"

	"bjoernblessin.de/go-utils/util/logger"
	"google.golang.org/grpc"
)

// Server represents the P2P gRPC server for peer-to-peer communication.
// This is the application stub / grpc adapter.
// It contains no domain logic, only type transformation and delegation.
type Server struct {
	pb.UnimplementedConnectionEstablishmentServer
	grpcServer        *grpc.Server
	listener          net.Listener
	connectionHandler handshake.HandshakeHandler
	peerRegistry      *peerregistry.PeerRegistry
}

func NewServer(handshakeHandler handshake.HandshakeHandler, peerRegistry *peerregistry.PeerRegistry) *Server {
	return &Server{
		connectionHandler: handshakeHandler,
		peerRegistry:      peerRegistry,
	}
}

// Start starts the P2P gRPC server on the given port in a goroutine.
func (s *Server) Start(port uint16) error {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	s.listener = listener

	s.grpcServer = grpc.NewServer()
	pb.RegisterConnectionEstablishmentServer(s.grpcServer, s)

	go func() {
		if err := s.grpcServer.Serve(listener); err != nil {
			logger.Warnf("gRPC server stopped with error: %v", err)
		}
	}()

	return nil
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
