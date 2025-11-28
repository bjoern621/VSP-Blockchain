package netzwerkroutinginterface

import (
	"context"
	"fmt"
	"net"
	"net/netip"

	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"

	"bjoernblessin.de/go-utils/util/logger"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Server represents the P2P gRPC server for peer-to-peer communication.
// This is the server skeleton (stub) in Tanenbaum's terminology.
type Server struct {
	pb.UnimplementedConnectionEstablishmentServer
	pb.UnimplementedPeerDiscoveryServer
	pb.UnimplementedErrorHandlingServer
	grpcServer *grpc.Server
	listener   net.Listener
}

// NewServer creates a new P2P server.
func NewServer() *Server {
	return &Server{}
}

// Start starts the P2P gRPC server on the given port.
func (s *Server) Start(port uint16) error {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	s.listener = listener

	s.grpcServer = grpc.NewServer()
	pb.RegisterConnectionEstablishmentServer(s.grpcServer, s)
	pb.RegisterPeerDiscoveryServer(s.grpcServer, s)
	pb.RegisterErrorHandlingServer(s.grpcServer, s)

	go func() {
		if err := s.grpcServer.Serve(listener); err != nil {
			// Log error but don't crash - server may have been stopped intentionally
		}
	}()

	return nil
}

// Stop gracefully stops the P2P server.
func (s *Server) Stop() {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
}

// --- ConnectionEstablishment service implementation ---

// Version handles incoming Version messages from peers initiating a connection.
func (s *Server) Version(ctx context.Context, req *pb.VersionInfo) (*emptypb.Empty, error) {
	// TODO: Implement proper version handling
	var listeningAddr string
	if req.ListeningEndpoint != nil {
		ip, ok := netip.AddrFromSlice(req.ListeningEndpoint.IpAddress)
		if ok {
			listeningAddr = netip.AddrPortFrom(ip, uint16(req.ListeningEndpoint.ListeningPort)).String()
		}
	}
	logger.Infof("Received Version from peer: version=%s, services=%v, listening=%s",
		req.Version, req.SupportedServices, listeningAddr)

	// 1. Validate version compatibility
	// 2. Store peer info
	// 3. Send Verack back to the peer
	return &emptypb.Empty{}, nil
}

// Verack handles incoming Verack messages acknowledging our Version message.
func (s *Server) Verack(ctx context.Context, req *pb.VersionInfo) (*emptypb.Empty, error) {
	// TODO: Implement proper verack handling
	// 1. Validate the verack
	// 2. Send Ack back to complete the handshake
	return &emptypb.Empty{}, nil
}

// Ack handles incoming Ack messages completing the connection establishment.
func (s *Server) Ack(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	// TODO: Implement proper ack handling
	// 1. Mark connection as fully established
	return &emptypb.Empty{}, nil
}

// --- PeerDiscovery service implementation ---

// GetAddr handles requests for known peer addresses.
func (s *Server) GetAddr(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	// TODO: Implement - respond with Addr containing known peers
	return &emptypb.Empty{}, nil
}

// Addr handles incoming peer address lists.
func (s *Server) Addr(ctx context.Context, req *pb.AddrList) (*emptypb.Empty, error) {
	// TODO: Implement - add new addresses to known peers list
	return &emptypb.Empty{}, nil
}

// Heartbeat handles keep-alive messages.
func (s *Server) Heartbeat(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	// TODO: Implement - update last_seen timestamp for the peer
	return &emptypb.Empty{}, nil
}

// --- ErrorHandling service implementation ---

// Reject handles error reports from peers.
func (s *Server) Reject(ctx context.Context, req *pb.Error) (*emptypb.Empty, error) {
	// TODO: Implement - log the error, possibly take corrective action
	return &emptypb.Empty{}, nil
}
