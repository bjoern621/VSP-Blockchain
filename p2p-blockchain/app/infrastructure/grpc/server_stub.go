package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"s3b/vsp-blockchain/p2p-blockchain/app/infrastructure/adapters"
	"s3b/vsp-blockchain/p2p-blockchain/wallet/api"

	"s3b/vsp-blockchain/p2p-blockchain/app/core"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"

	"bjoernblessin.de/go-utils/util/logger"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
)

// Server represents the app gRPC server for external local systems.
type Server struct {
	pb.UnimplementedAppServiceServer
	grpcServer           *grpc.Server
	listener             net.Listener
	connService          *core.ConnectionEstablishmentService
	regService           *core.InternalViewService
	queryRegistryService *core.QueryRegistryService
	discoveryService     *core.DiscoveryService
	keysApi              api.KeyGeneratorApi
	transactionHandler   *adapters.TransactionHandlerAdapter
	kontoHandler         *adapters.KontoHandlerAdapter
	visualizationHandler *adapters.VisualizationHandlerAdapter
}

// NewServer creates a new external API server.
func NewServer(
	connService *core.ConnectionEstablishmentService,
	regService *core.InternalViewService,
	queryRegistryService *core.QueryRegistryService,
	keysApi api.KeyGeneratorApi,
	transactionHandler *adapters.TransactionHandlerAdapter,
	discoveryService *core.DiscoveryService,
	kontoHandler *adapters.KontoHandlerAdapter,
	visualizationHandler *adapters.VisualizationHandlerAdapter,
) *Server {
	return &Server{
		connService:          connService,
		regService:           regService,
		queryRegistryService: queryRegistryService,
		discoveryService:     discoveryService,
		keysApi:              keysApi,
		transactionHandler:   transactionHandler,
		kontoHandler:         kontoHandler,
		visualizationHandler: visualizationHandler,
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
			logger.Warnf("[app server] App gRPC server stopped with error: %v", err)
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
func (s *Server) ConnectTo(_ context.Context, req *pb.ConnectToRequest) (*pb.ConnectToResponse, error) {
	if req.Port > 65535 {
		return &pb.ConnectToResponse{
			Success:      false,
			ErrorMessage: "port must be between 0 and 65535",
		}, nil
	}

	port := uint16(req.Port)

	err := s.connService.ConnectTo(req.IpAddress, port)
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
func (s *Server) GetInternalPeerInfo(_ context.Context, _ *pb.GetInternalPeerInfoRequest) (*pb.GetInternalPeerInfoResponse, error) {
	peers := s.regService.GetInternalPeerInfo()

	response := &pb.GetInternalPeerInfoResponse{
		Entries: make([]*pb.InternalPeerInfoEntry, 0, len(peers)),
	}

	for _, p := range peers {
		infraStruct, err := structpb.NewStruct(p.PeerInfrastructureData)
		if err != nil {
			logger.Warnf("[app server] failed to create structpb from infra data: %v", err)
			infraStruct, _ = structpb.NewStruct(nil)
		}

		entry := &pb.InternalPeerInfoEntry{
			PeerId:             string(p.PeerID),
			InfrastructureData: infraStruct,
			Version:            p.Version,
			ConnectionState:    p.ConnectionState.String(),
			LastSeen:           p.LastSeen,
		}

		for _, svc := range p.SupportedServices {
			entry.SupportedServices = append(entry.SupportedServices, svc.String())
		}

		response.Entries = append(response.Entries, entry)
	}

	return response, nil
}

// QueryRegistry queries the DNS seed registry for available peer addresses.
func (s *Server) QueryRegistry(_ context.Context, _ *pb.QueryRegistryRequest) (*pb.QueryRegistryResponse, error) {
	entries, err := s.queryRegistryService.QueryRegistry()
	if err != nil {
		return nil, err
	}

	response := &pb.QueryRegistryResponse{
		Entries: make([]*pb.RegistryEntry, 0, len(entries)),
	}

	for _, entry := range entries {
		response.Entries = append(response.Entries, &pb.RegistryEntry{
			IpAddress: entry.IPAddress.String(),
			PeerId:    string(entry.PeerID),
		})
	}

	return response, nil
}
func (s *Server) CreateTransaction(_ context.Context, req *pb.CreateTransactionRequest) (*pb.CreateTransactionResponse, error) {
	return s.transactionHandler.CreateTransaction(req), nil
}

func (s *Server) GenerateKeyset(context.Context, *emptypb.Empty) (*pb.GenerateKeysetResponse, error) {
	keyset := s.keysApi.GenerateKeyset()

	pbKeyset := pb.Keyset{
		PrivateKey:    keyset.PrivateKey[:],
		PrivateKeyWif: keyset.PrivateKeyWif,
		PublicKey:     keyset.PublicKey[:],
		VSAddress:     keyset.VSAddress,
	}

	response := &pb.GenerateKeysetResponse{
		Keyset: &pbKeyset,
	}
	return response, nil
}

func (s *Server) GetKeysetFromWIF(_ context.Context, wif *pb.GetKeysetFromWIFRequest) (*pb.GetKeysetFromWIFResponse, error) {
	keyset, err := s.keysApi.GetKeysetFromWIF(wif.PrivateKeyWif)

	if err != nil {
		return &pb.GetKeysetFromWIFResponse{
			FalseInput: true,
		}, nil
	}

	pbKeyset := pb.Keyset{
		PrivateKey:    keyset.PrivateKey[:],
		PrivateKeyWif: keyset.PrivateKeyWif,
		PublicKey:     keyset.PublicKey[:],
		VSAddress:     keyset.VSAddress,
	}

	response := &pb.GetKeysetFromWIFResponse{
		Keyset:     &pbKeyset,
		FalseInput: false,
	}
	return response, nil
}

// SendGetAddr handles the SendGetAddr RPC call from external local systems.
// This allows manual triggering of getaddr requests to specific peers.
func (s *Server) SendGetAddr(_ context.Context, req *pb.SendGetAddrRequest) (*pb.SendGetAddrResponse, error) {
	err := s.discoveryService.SendGetAddr(req.PeerId)
	if err != nil {
		return &pb.SendGetAddrResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to send GetAddr: %v", err),
		}, nil
	}

	return &pb.SendGetAddrResponse{
		Success:      true,
		ErrorMessage: "",
	}, nil
}

func (s *Server) GetAssets(_ context.Context, req *pb.GetAssetsRequest) (*pb.GetAssetsResponse, error) {
	return s.kontoHandler.GetAssets(req), nil
}

func (s *Server) GetBlockchainVisualization(_ context.Context, req *pb.GetBlockchainVisualizationRequest) (*pb.GetBlockchainVisualizationResponse, error) {
	return s.visualizationHandler.GetBlockchainVisualization(req), nil
}
