package core

import (
	"context"
	"net/netip"
)

// VersionInfo represents version information exchanged during handshake.
// This is a domain model, independent of the transport layer (gRPC/protobuf).
type VersionInfo struct {
	Version           string
	SupportedServices []string
	ListeningEndpoint netip.AddrPort
}

// HandshakeHandler defines the interface for handling incoming connection messages.
// This interface is implemented in the core/domain layer and used by the infrastructure layer.
type HandshakeHandler interface {
	HandleVersion(ctx context.Context, peerAddr string, info VersionInfo) error
	HandleVerack(ctx context.Context, peerAddr string, info VersionInfo) error
	HandleAck(ctx context.Context, peerAddr string) error
}

// HandshakeService implements ConnectionHandler with the actual domain logic.
type HandshakeService struct {
	// Add dependencies here (e.g., peer store, message sender)
}

func NewHandshakeService() *HandshakeService {
	return &HandshakeService{}
}

func (h *HandshakeService) HandleVersion(ctx context.Context, peerAddr string, info VersionInfo) error {
	// Domain logic:
	// 1. Validate version compatibility
	// 2. Store peer info
	// 3. Send Verack back to the peer (via MessageSender interface)
	return nil
}

func (h *HandshakeService) HandleVerack(ctx context.Context, peerAddr string, info VersionInfo) error {
	// Domain logic:
	// 1. Validate the verack
	// 2. Send Ack back to complete the handshake
	return nil
}

func (h *HandshakeService) HandleAck(ctx context.Context, peerAddr string) error {
	// Domain logic:
	// 1. Mark connection as fully established
	return nil
}

func InitiateHandshake() {
	// Peer A initiates the handshake by sending a Version message to Peer B.
	// Peer B responds with a Verack message.
	// Peer A then sends an Ack message to complete the handshake.
}
