package core

import (
	"fmt"
	"net/netip"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type VersionInfo struct {
	Version           string
	SupportedServices []string
	ListeningEndpoint netip.AddrPort
}

// HandshakeHandler defines the interface for handling incoming connection messages.
// This interface is implemented in the core/domain layer and used by the infrastructure layer.
type HandshakeHandler interface {
	HandleVersion(peerID peer.PeerID, info VersionInfo)
	HandleVerack(peerID peer.PeerID, info VersionInfo)
	HandleAck(peerID peer.PeerID)
}

// HandshakeService implements ConnectionHandler with the actual domain logic.
type HandshakeService struct {
	// Add dependencies here (e.g., peer store, message sender)
}

func NewHandshakeService() *HandshakeService {
	return &HandshakeService{}
}

func (h *HandshakeService) HandleVersion(peerID peer.PeerID, info VersionInfo) {
	// Domain logic:
	// 1. Validate version compatibility
	// 2. Store peer info
	// 3. Send Verack back to the peer (via MessageSender interface)
}

func (h *HandshakeService) HandleVerack(peerID peer.PeerID, info VersionInfo) {
	// Domain logic:
	// 1. Validate the verack
	// 2. Send Ack back to complete the handshake
}

func (h *HandshakeService) HandleAck(peerID peer.PeerID) {
	// Domain logic:
	// 1. Mark connection as fully established
}

func InitiateHandshake(addrPort netip.AddrPort) error {
	addr := addrPort.String()

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", addr, err)
	}
	defer conn.Close()

	client := pb.NewConnectionEstablishmentClient(conn)

	versionInfo := &pb.VersionInfo{
		Version:           "vsgoin-1.0",
		SupportedServices: []pb.ServiceType{pb.ServiceType_SERVICE_NETZWERKROUTING, pb.ServiceType_SERVICE_BLOCKCHAIN_FULL, pb.ServiceType_SERVICE_WALLET, pb.ServiceType_SERVICE_MINER},
		ListeningEndpoint: &pb.Endpoint{
			IpAddress:     common.P2PListeningIpAddr.AsSlice(),
			ListeningPort: uint32(common.P2PPort),
		},
	}

	_, err = client.Version(ctx, versionInfo)
	if err != nil {
		return fmt.Errorf("failed to send version: %w", err)
	}

	return nil
}
