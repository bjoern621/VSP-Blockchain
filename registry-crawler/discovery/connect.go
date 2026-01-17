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
		logger.Debugf("[discovery] ConnectTo %s:%d failed: %v", ip, port, connectErr)
		return false, connectErr
	}

	if resp != nil && resp.Success {
		logger.Debugf("[discovery] ConnectTo %s:%d result: success=true", ip, port)
		return true, nil
	}

	errorMsg := ""
	if resp != nil {
		errorMsg = strings.TrimSpace(resp.ErrorMessage)
	}

	if isAlreadyConnectedError(errorMsg) {
		logger.Debugf("[discovery] ConnectTo %s:%d result: peer already connected (treating as success)", ip, port)
		return true, nil
	}

	logger.Debugf("[discovery] ConnectTo %s:%d result: success=false reason=%q", ip, port, errorMsg)
	return false, fmt.Errorf("connect_to failed: %s", errorMsg)
}

// isAlreadyConnectedError checks if the error message indicates the peer is already connected.
func isAlreadyConnectedError(errorMsg string) bool {
	return strings.Contains(errorMsg, "state connected") ||
		strings.Contains(errorMsg, "already connected")
}
