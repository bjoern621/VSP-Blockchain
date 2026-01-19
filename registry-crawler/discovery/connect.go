package discovery

import (
	"context"
	"fmt"
	"net/netip"
	"s3b/vsp-blockchain/registry-crawler/internal/pb"
	"strings"

	"bjoernblessin.de/go-utils/util/logger"
)

// ConnectToPeer attempts to connect to a single peer via the app service.
// Returns true if the connection succeeded or the peer is already connected.
func ConnectToPeer(ctx context.Context, client pb.AppServiceClient, ip string, port int32) (success bool, err error) {
	parsed, parseErr := netip.ParseAddr(ip)
	if parseErr != nil {
		return false, fmt.Errorf("invalid peer IP %s: %w", ip, parseErr)
	}

	resp, connectErr := client.ConnectTo(ctx, &pb.ConnectToRequest{
		IpAddress: parsed.AsSlice(),
		Port:      uint32(port),
	})
	if connectErr != nil {
		logger.Debugf("ConnectTo %s:%d failed: %v", ip, port, connectErr)
		return false, connectErr
	}

	if resp != nil && resp.Success {
		logger.Debugf("ConnectTo %s:%d result: success=true", ip, port)
		return true, nil
	}

	errorMsg := ""
	if resp != nil {
		errorMsg = strings.TrimSpace(resp.ErrorMessage)
	}

	if isAlreadyConnectedError(errorMsg) {
		logger.Debugf("ConnectTo %s:%d result: peer already connected (treating as success)", ip, port)
		return true, nil
	}

	logger.Debugf("ConnectTo %s:%d result: success=false reason=%q", ip, port, errorMsg)
	return false, fmt.Errorf("connect_to failed: %s", errorMsg)
}

// DisconnectPeer attempts to disconnect from a single peer via the app service.
// Disconnecting means forgetting the peer (removing it from peer store and network info registry).
// Returns true if the disconnection succeeded or the peer was already not connected.
func DisconnectPeer(ctx context.Context, client pb.AppServiceClient, ip string, port int32) (success bool, err error) {
	parsed, parseErr := netip.ParseAddr(ip)
	if parseErr != nil {
		return false, fmt.Errorf("invalid peer IP %s: %w", ip, parseErr)
	}

	resp, disconnectErr := client.Disconnect(ctx, &pb.DisconnectRequest{
		IpAddress: parsed.AsSlice(),
		Port:      uint32(port),
	})
	if disconnectErr != nil {
		logger.Debugf("Disconnect %s:%d failed: %v", ip, port, disconnectErr)
		return false, disconnectErr
	}

	if resp != nil && resp.Success {
		logger.Debugf("Disconnect %s:%d result: success=true", ip, port)
		return true, nil
	}

	errorMsg := ""
	if resp != nil {
		errorMsg = strings.TrimSpace(resp.ErrorMessage)
	}

	// If the peer wasn't found, we can consider that a success
	if isNotFoundError(errorMsg) {
		logger.Debugf("Disconnect %s:%d result: peer not found (treating as success)", ip, port)
		return true, nil
	}

	logger.Debugf("Disconnect %s:%d result: success=false reason=%q", ip, port, errorMsg)
	return false, fmt.Errorf("disconnect failed: %s", errorMsg)
}

// isAlreadyConnectedError checks if the error message indicates the peer is already connected.
func isAlreadyConnectedError(errorMsg string) bool {
	return strings.Contains(errorMsg, "state connected") ||
		strings.Contains(errorMsg, "already connected")
}

// isNotFoundError checks if the error message indicates the peer was not found.
func isNotFoundError(errorMsg string) bool {
	return strings.Contains(errorMsg, "peer not found") ||
		strings.Contains(errorMsg, "not found")
}

// isHolddownError checks if the error message indicates the peer is in holddown state.
// Peers in holddown were recently disconnected and reject new connection attempts.
func isHolddownError(errorMsg string) bool {
	return strings.Contains(errorMsg, "holddown") ||
		strings.Contains(errorMsg, "recently disconnected")
}
