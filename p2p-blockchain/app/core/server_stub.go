package core

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"

	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/infrastructure/middleware/grpc/networkinfo"

	"bjoernblessin.de/go-utils/util/logger"
	"google.golang.org/grpc"
)

// Server represents the app gRPC server for external local systems.
type Server struct {
	pb.UnimplementedAppServiceServer
	grpcServer          *grpc.Server
	listener            net.Listener
	handshakeAPI        api.HandshakeAPI
	networkInfoRegistry *networkinfo.NetworkInfoRegistry
	peerStore           *peer.PeerStore
}

// NewServer creates a new external API server.
func NewServer(handshakeAPI api.HandshakeAPI, networkInfoRegistry *networkinfo.NetworkInfoRegistry, peerStore *peer.PeerStore) *Server {
	return &Server{
		handshakeAPI:        handshakeAPI,
		networkInfoRegistry: networkInfoRegistry,
		peerStore:           peerStore,
	}
}

// Start starts the external API gRPC server on the given port in a goroutine.
func (s *Server) Start(port uint16) error {
	addr := fmt.Sprintf("%s:%d", common.AppListenAddr(), port)
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

	addrPort := netip.AddrPortFrom(ip, port)

	err := s.handshakeAPI.InitiateHandshake(addrPort)
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
func (s *Server) GetPeerRegistry(ctx context.Context, req *pb.GetPeerRegistryRequest) (*pb.GetPeerRegistryResponse, error) {
	allInfo := s.networkInfoRegistry.GetAllNetworkInfo()

	response := &pb.GetPeerRegistryResponse{
		Entries: make([]*pb.PeerRegistryEntry, 0, len(allInfo)),
	}

	for _, info := range allInfo {
		entry := &pb.PeerRegistryEntry{
			PeerId:                string(info.PeerID),
			HasOutboundConnection: info.HasOutboundConn,
		}

		// Listening endpoint
		if info.ListeningEndpoint.IsValid() {
			entry.ListeningEndpoint = &pb.Endpoint{
				IpAddress:     info.ListeningEndpoint.Addr().AsSlice(),
				ListeningPort: uint32(info.ListeningEndpoint.Port()),
			}
		}

		// Inbound addresses
		for _, addr := range info.InboundAddresses {
			entry.InboundAddresses = append(entry.InboundAddresses, &pb.Endpoint{
				IpAddress:     addr.Addr().AsSlice(),
				ListeningPort: uint32(addr.Port()),
			})
		}

		// Peer info from PeerStore
		if p, exists := s.peerStore.GetPeer(info.PeerID); exists {
			p.Lock()
			entry.Version = p.Version
			entry.ConnectionState = connectionStateToString(p.State)
			entry.Direction = directionToString(p.Direction)

			for _, svc := range p.SupportedServices {
				entry.SupportedServices = append(entry.SupportedServices, serviceTypeToString(svc))
			}
			p.Unlock()
		}

		response.Entries = append(response.Entries, entry)
	}

	return response, nil
}

func connectionStateToString(state peer.PeerConnectionState) string {
	switch state {
	case peer.StateNew:
		return "new"
	case peer.StateAwaitingVerack:
		return "awaiting_verack"
	case peer.StateAwaitingAck:
		return "awaiting_ack"
	case peer.StateConnected:
		return "connected"
	default:
		return "unknown"
	}
}

func directionToString(dir peer.Direction) string {
	switch dir {
	case peer.DirectionInbound:
		return "inbound"
	case peer.DirectionOutbound:
		return "outbound"
	case peer.DirectionBoth:
		return "both"
	default:
		return "unknown"
	}
}

func serviceTypeToString(svc peer.ServiceType) string {
	switch svc {
	case peer.ServiceType_Netzwerkrouting:
		return "netzwerkrouting"
	case peer.ServiceType_BlockchainFull:
		return "blockchain_full"
	case peer.ServiceType_BlockchainSimple:
		return "blockchain_simple"
	case peer.ServiceType_Wallet:
		return "wallet"
	case peer.ServiceType_Miner:
		return "miner"
	default:
		return "unknown"
	}
}
