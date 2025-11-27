package core

import (
	"context"
	"fmt"
	"net/netip"

	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ConnectTo initiates a connection to a peer at the given IP and port.
// It sends a Version message to start the 3-way handshake.
func ConnectTo(ctx context.Context, ip netip.Addr, port uint16) error {
	addrPort := netip.AddrPortFrom(ip, port)
	addr := addrPort.String()

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", addr, err)
	}
	defer conn.Close()

	client := pb.NewConnectionEstablishmentClient(conn)

	versionInfo := &pb.VersionInfo{
		Version:           "vsgoin-1.0",
		SupportedServices: []pb.ServiceType{pb.ServiceType_SERVICE_NETZWERKROUTING},
		ListeningEndpoint: &pb.Endpoint{
			IpAddress:     []byte{0, 0, 0, 0}, // TODO: set actual local IP
			ListeningPort: 50051,              // TODO: set actual local port
		},
	}

	_, err = client.Version(ctx, versionInfo)
	if err != nil {
		return fmt.Errorf("failed to send version: %w", err)
	}

	return nil
}
