package main

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"slices"
	"strconv"
	"strings"
	"sync"

	"s3b/vsp-blockchain/registry-crawler/internal/pb"

	"bjoernblessin.de/go-utils/util/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// resolveBootstrapEndpoints resolves bootstrap endpoints to IP addresses.
// Returns a set of IP addresses and the P2P port.
func resolveBootstrapEndpoints(ctx context.Context, cfg Config) (map[string]struct{}, int32) {
	res := map[string]struct{}{}

	endpoints := make([]string, 0, len(cfg.Bootstrap.Endpoints))
	endpoints = append(endpoints, cfg.Bootstrap.Endpoints...)

	for _, token := range endpoints {
		host, port, err := splitHostPortOrDefault(token, int(cfg.AcceptedP2PPort))
		if err != nil {
			continue
		}
		if uint16(port) != cfg.AcceptedP2PPort {
			continue
		}

		if parsed, err := netip.ParseAddr(host); err == nil {
			if parsed.Is4() {
				res[parsed.String()] = struct{}{}
			}
			continue
		}

		resolved, err := net.DefaultResolver.LookupIPAddr(ctx, host)
		if err != nil {
			continue
		}
		for _, r := range resolved {
			if r.IP == nil {
				continue
			}
			addr, ok := netip.AddrFromSlice(r.IP)
			if !ok {
				continue
			}
			if addr.Is4() {
				res[addr.String()] = struct{}{}
			}
		}
	}

	return res, int32(cfg.AcceptedP2PPort)
}

// fetchNetworkPeers queries the app service for connected peers from the P2P network.
// Returns a set of IP addresses and the P2P port.
func fetchNetworkPeers(ctx context.Context, cfg Config) (map[string]struct{}, int32, error) {
	conn, err := dialAppGRPC(ctx, cfg.AppAddr)
	if err != nil {
		return nil, 0, err
	}

	client := pb.NewAppServiceClient(conn)

	resp, err := client.GetInternalPeerInfo(ctx, &pb.GetInternalPeerInfoRequest{})
	if err != nil {
		return nil, int32(cfg.AcceptedP2PPort), nil
	}

	ips := map[string]struct{}{}
	acceptedPort := uint16(cfg.AcceptedP2PPort)

	for _, entry := range resp.GetEntries() {
		if entry == nil {
			continue
		}

		if !slices.Contains(entry.SupportedServices, "miner") {
			continue
		}
		if entry.ConnectionState != "connected" {
			continue
		}

		infra := entry.GetInfrastructureData()
		if infra == nil {
			continue
		}
		infraMap := infra.AsMap()
		listeningEndpoint, ok := infraMap["listeningEndpoint"].(string)
		if !ok {
			continue
		}
		ap, err := netip.ParseAddrPort(strings.TrimSpace(listeningEndpoint))
		if err != nil {
			continue
		}
		if ap.Port() != acceptedPort {
			continue
		}

		addr := ap.Addr()
		if !addr.Is4() {
			continue
		}
		ips[addr.String()] = struct{}{}
	}

	return ips, int32(cfg.AcceptedP2PPort), nil
}

// connectToPeer attempts to connect to a single peer via the app service.
// Returns true if the connection succeeded or the peer is already connected.
func connectToPeer(ctx context.Context, client pb.AppServiceClient, ip string, port int32) (success bool, err error) {
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

// isAlreadyConnectedError checks if the error message indicates the peer is already connected.
func isAlreadyConnectedError(errorMsg string) bool {
	return strings.Contains(errorMsg, "state connected") ||
		strings.Contains(errorMsg, "already connected")
}

// splitHostPortOrDefault splits a host:port string or uses the default port.
func splitHostPortOrDefault(token string, defaultPort int) (string, int, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return "", 0, fmt.Errorf("empty endpoint")
	}

	host, portString, err := net.SplitHostPort(token)
	if err != nil {
		return token, defaultPort, nil
	}
	if host == "" {
		return "", 0, fmt.Errorf("empty host")
	}
	port, err := strconv.Atoi(portString)
	if err != nil {
		return "", 0, err
	}
	return host, port, nil
}

var (
	appGRPCConnMu   sync.Mutex
	appGRPCConn     *grpc.ClientConn
	appGRPCConnAddr string
)

// dialAppGRPC establishes a gRPC connection to the app service.
func dialAppGRPC(ctx context.Context, addr string) (*grpc.ClientConn, error) {
	_ = ctx

	appGRPCConnMu.Lock()
	defer appGRPCConnMu.Unlock()

	if appGRPCConn != nil {
		if appGRPCConnAddr != addr {
			return nil, fmt.Errorf("app grpc connection already initialized for %q; cannot reuse for %q", appGRPCConnAddr, addr)
		}
		return appGRPCConn, nil
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	appGRPCConn = conn
	appGRPCConnAddr = addr
	return appGRPCConn, nil
}
