package grpc

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"s3b/vsp-blockchain/p2p-blockchain/app/infrastructure/adapters"
	blockcahin_api "s3b/vsp-blockchain/p2p-blockchain/blockchain/api"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
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
	disconnectService    *core.DisconnectService
	keysApi              api.KeyGeneratorApi
	transactionAPI       api.TransactionCreationAPI
	kontoAPI             api.KontoAPI
	historyAPI           api.HistoryAPI
	visualizationHandler *adapters.VisualizationHandlerAdapter
	miningService        *core.MiningService
	blockStore           blockcahin_api.BlockStoreAPI
}

// NewServer creates a new external API server.
func NewServer(
	connService *core.ConnectionEstablishmentService,
	regService *core.InternalViewService,
	queryRegistryService *core.QueryRegistryService,
	keysApi api.KeyGeneratorApi,
	transactionAPI api.TransactionCreationAPI,
	discoveryService *core.DiscoveryService,
	kontoAPI api.KontoAPI,
	historyAPI api.HistoryAPI,
	visualizationHandler *adapters.VisualizationHandlerAdapter,
	miningService *core.MiningService,
	disconnectService *core.DisconnectService,
	blockStore blockcahin_api.BlockStoreAPI,
) *Server {
	return &Server{
		connService:          connService,
		regService:           regService,
		queryRegistryService: queryRegistryService,
		discoveryService:     discoveryService,
		disconnectService:    disconnectService,
		keysApi:              keysApi,
		transactionAPI:       transactionAPI,
		kontoAPI:             kontoAPI,
		historyAPI:           historyAPI,
		visualizationHandler: visualizationHandler,
		miningService:        miningService,
		blockStore:           blockStore,
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

// Disconnect handles the Disconnect RPC call from external local systems.
// Disconnecting from a peer means forgetting the peer, which involves:
// - Closing any gRPC connections
// - Removing the peer from network info registry
// - Removing the peer from peer store
func (s *Server) Disconnect(_ context.Context, req *pb.DisconnectRequest) (*pb.DisconnectResponse, error) {
	if req.Port > 65535 {
		return &pb.DisconnectResponse{
			Success:      false,
			ErrorMessage: "port must be between 0 and 65535",
		}, nil
	}

	port := uint16(req.Port)

	err := s.disconnectService.Disconnect(req.IpAddress, port)
	if err != nil {
		return &pb.DisconnectResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to disconnect: %v", err),
		}, nil
	}

	return &pb.DisconnectResponse{
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
	if s.transactionAPI == nil {
		return &pb.CreateTransactionResponse{
			Success:      false,
			ErrorCode:    pb.TransactionErrorCode_VALIDATION_FAILED,
			ErrorMessage: "wallet subsystem is not enabled",
		}, nil
	}

	// Validate request fields
	if req.RecipientVsAddress == "" {
		return &pb.CreateTransactionResponse{
			Success:      false,
			ErrorCode:    pb.TransactionErrorCode_VALIDATION_FAILED,
			ErrorMessage: "recipient address is required",
		}, nil
	}
	if req.Amount == 0 {
		return &pb.CreateTransactionResponse{
			Success:      false,
			ErrorCode:    pb.TransactionErrorCode_VALIDATION_FAILED,
			ErrorMessage: "amount must be greater than 0",
		}, nil
	}
	if req.SenderPrivateKeyWif == "" {
		return &pb.CreateTransactionResponse{
			Success:      false,
			ErrorCode:    pb.TransactionErrorCode_INVALID_PRIVATE_KEY,
			ErrorMessage: "sender private key is required",
		}, nil
	}

	result := s.transactionAPI.CreateTransaction(req.RecipientVsAddress, req.Amount, req.SenderPrivateKeyWif)

	// Map error code
	var pbErrorCode pb.TransactionErrorCode
	switch result.ErrorCode {
	case transaction.ErrorCodeNone:
		pbErrorCode = pb.TransactionErrorCode_NONE
	case transaction.ErrorCodeInvalidPrivateKey:
		pbErrorCode = pb.TransactionErrorCode_INVALID_PRIVATE_KEY
	case transaction.ErrorCodeInsufficientFunds:
		pbErrorCode = pb.TransactionErrorCode_INSUFFICIENT_FUNDS
	case transaction.ErrorCodeValidationFailed:
		pbErrorCode = pb.TransactionErrorCode_VALIDATION_FAILED
	case transaction.ErrorCodeBroadcastFailed:
		pbErrorCode = pb.TransactionErrorCode_BROADCAST_FAILED
	default:
		pbErrorCode = pb.TransactionErrorCode_VALIDATION_FAILED
	}

	return &pb.CreateTransactionResponse{
		Success:       result.Success,
		ErrorCode:     pbErrorCode,
		ErrorMessage:  result.ErrorMessage,
		TransactionId: result.TransactionID,
	}, nil
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
	if s.kontoAPI == nil {
		return &pb.GetAssetsResponse{
			Success:      false,
			ErrorMessage: "wallet subsystem is not enabled",
		}, nil
	}

	if req.VsAddress == "" {
		return &pb.GetAssetsResponse{
			Success:      false,
			ErrorMessage: "V$Address is required",
		}, nil
	}

	result := s.kontoAPI.GetAssets(req.VsAddress)

	if !result.Success {
		return &pb.GetAssetsResponse{
			Success:      false,
			ErrorMessage: result.ErrorMessage,
		}, nil
	}

	// Convert assets to protobuf format
	pbAssets := make([]*pb.Asset, 0, len(result.Assets))
	for _, asset := range result.Assets {
		pbAssets = append(pbAssets, &pb.Asset{
			Value: asset.Value,
		})
	}

	return &pb.GetAssetsResponse{
		Success: true,
		Assets:  pbAssets,
	}, nil
}

func (s *Server) GetHistory(_ context.Context, req *pb.GetHistoryRequest) (*pb.GetHistoryResponse, error) {
	if s.historyAPI == nil {
		return &pb.GetHistoryResponse{
			Success:      false,
			ErrorMessage: "wallet subsystem is not enabled",
		}, nil
	}

	if req.VsAddress == "" {
		return &pb.GetHistoryResponse{
			Success:      false,
			ErrorMessage: "V$Address is required",
		}, nil
	}

	result := s.historyAPI.GetHistory(req.VsAddress)

	if !result.Success {
		return &pb.GetHistoryResponse{
			Success:      false,
			ErrorMessage: result.ErrorMessage,
		}, nil
	}

	// Convert transactions to string format for response
	txStrings := make([]string, 0, len(result.Transactions))
	for _, tx := range result.Transactions {
		var txStr string
		if tx.IsSender && tx.Received > 0 {
			txStr = fmt.Sprintf("TxID: %s, Block: %d, Sent: %d, Received: %d",
				tx.TransactionID, tx.BlockHeight, tx.Sent, tx.Received)
		} else if tx.IsSender {
			txStr = fmt.Sprintf("TxID: %s, Block: %d, Sent: %d",
				tx.TransactionID, tx.BlockHeight, tx.Sent)
		} else if tx.Received > 0 {
			txStr = fmt.Sprintf("TxID: %s, Block: %d, Received: %d",
				tx.TransactionID, tx.BlockHeight, tx.Received)
		} else {
			logger.Warnf("TransactionEntry with zero sent and received amounts: %+v", tx)
		}
		txStrings = append(txStrings, txStr)
	}

	return &pb.GetHistoryResponse{
		Success:      true,
		Transactions: txStrings,
	}, nil
}

func (s *Server) GetBlockchainVisualization(_ context.Context, req *pb.GetBlockchainVisualizationRequest) (*pb.GetBlockchainVisualizationResponse, error) {
	return s.visualizationHandler.GetBlockchainVisualization(req), nil
}

func (s *Server) StartMining(_ context.Context, _ *emptypb.Empty) (*pb.StartMiningResponse, error) {
	if s.miningService == nil {
		return &pb.StartMiningResponse{
			Success:      false,
			ErrorMessage: "mining subsystem is not enabled",
		}, nil
	}

	err := s.miningService.EnableMining()
	if err != nil {
		return &pb.StartMiningResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to enable mining: %v", err),
		}, nil
	}

	// Attempt to start mining with empty transactions
	err = s.miningService.StartMining(nil)
	if err != nil {
		return &pb.StartMiningResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to start mining: %v", err),
		}, nil
	}

	return &pb.StartMiningResponse{
		Success:      true,
		ErrorMessage: "",
	}, nil
}

func (s *Server) StopMining(_ context.Context, _ *emptypb.Empty) (*pb.StopMiningResponse, error) {
	if s.miningService == nil {
		return &pb.StopMiningResponse{
			Success:      false,
			ErrorMessage: "mining subsystem is not enabled",
		}, nil
	}

	err := s.miningService.DisableMining()
	if err != nil {
		return &pb.StopMiningResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to disable mining: %v", err),
		}, nil
	}

	return &pb.StopMiningResponse{
		Success:      true,
		ErrorMessage: "",
	}, nil
}

func (s *Server) GetConfirmationStatus(_ context.Context, request *pb.GetConfirmationStatusRequest) (*pb.GetConfirmationStatusResponse, error) {
	txIDString := request.GetTransactionId()

	txIDSlice, err := hex.DecodeString(txIDString)
	if err != nil {
		return &pb.GetConfirmationStatusResponse{Accepted: false}, nil
	}

	var txID transaction.TransactionID

	if len(txIDSlice) != len(txID) {
		return &pb.GetConfirmationStatusResponse{Accepted: false}, nil
	}

	copy(txID[:], txIDSlice)

	result, err := s.blockStore.IsTransactionAccepted(txID)

	if err != nil {
		return &pb.GetConfirmationStatusResponse{Accepted: false}, nil
	}

	return &pb.GetConfirmationStatusResponse{Accepted: result}, nil
}
