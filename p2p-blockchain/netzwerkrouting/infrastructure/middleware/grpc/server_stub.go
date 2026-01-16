package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"

	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api/blockchain/observer"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/handshake"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/keepalive"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer/discovery"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/infrastructure/middleware/grpc/networkinfo"

	"bjoernblessin.de/go-utils/util/logger"
	mapset "github.com/deckarep/golang-set/v2"
	"google.golang.org/grpc"
)

// Server represents the P2P gRPC server for peer-to-peer communication.
// This is the application stub / grpc adapter.
// It contains no domain logic, only type transformation and delegation.
type Server struct {
	pb.UnimplementedConnectionEstablishmentServer
	grpcServer          *grpc.Server
	listener            net.Listener
	handshakeMsgHandler handshake.HandshakeMsgHandler
	networkInfoRegistry *networkinfo.NetworkInfoRegistry

	pb.UnimplementedBlockchainServiceServer
	observers mapset.Set[observer.BlockchainObserverAPI]

	pb.UnimplementedPeerDiscoveryServer
	discoveryService *discovery.DiscoveryService
	keepaliveService *keepalive.KeepaliveService
}

func NewServer(handshakeMsgHandler handshake.HandshakeMsgHandler, networkInfoRegistry *networkinfo.NetworkInfoRegistry, discoveryService *discovery.DiscoveryService, keepaliveService *keepalive.KeepaliveService) *Server {
	return &Server{
		handshakeMsgHandler: handshakeMsgHandler,
		networkInfoRegistry: networkInfoRegistry,
		observers:           mapset.NewSet[observer.BlockchainObserverAPI](),
		discoveryService:    discoveryService,
		keepaliveService:    keepaliveService,
	}
}

// Start starts the P2P gRPC server on the given port in a goroutine.
func (s *Server) Start(port uint16) error {
	addr := fmt.Sprintf("%s:%d", common.P2PListenAddr(), port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	s.listener = listener

	s.grpcServer = grpc.NewServer()
	pb.RegisterConnectionEstablishmentServer(s.grpcServer, s)
	pb.RegisterBlockchainServiceServer(s.grpcServer, s)
	pb.RegisterPeerDiscoveryServer(s.grpcServer, s)

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

func (s *Server) Attach(o observer.BlockchainObserverAPI) {
	s.observers.Add(o)
}

func (s *Server) Detach(o observer.BlockchainObserverAPI) {
	s.observers.Remove(o)
}

// GetPeerId retrieves the PeerId associated with the incoming gRPC context.
// It registers the peer if it is not already known.
// The registed / created peer will be available in data layers PeerStore.
// The peer will have no direction assigned yet.
func (s *Server) GetPeerId(ctx context.Context) common.PeerId {
	inboundAddr := getPeerAddr(ctx)
	peerID := s.networkInfoRegistry.GetOrRegisterPeer(inboundAddr, netip.AddrPort{})
	s.networkInfoRegistry.AddInboundAddress(peerID, inboundAddr)

	return peerID
}
