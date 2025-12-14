package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"strings"

	"s3b/vsp-blockchain/registry-crawler/internal/pb"

	"bjoernblessin.de/go-utils/util/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// fetchSeedTargets discovers node IPs that have blockchain_full subsystem from the P2P network.
// Returns a set of IP addresses and the P2P port. Uses bootstrap endpoints
// for initial connectivity, then queries the app service for connected peers.
func fetchSeedTargets(ctx context.Context, cfg Config) (map[string]struct{}, int32, error) {
	bootstrapTargets := parseBootstrapTargets(ctx, cfg)

	if len(cfg.OverrideIPs) > 0 {
		ips := map[string]struct{}{}
		for _, ipString := range cfg.OverrideIPs {
			ips[ipString] = struct{}{}
		}
		for ip := range bootstrapTargets {
			ips[ip] = struct{}{}
		}
		return ips, int32(cfg.P2PPort), nil
	}

	conn, err := dialAppGRPC(ctx, cfg.AppAddr, false)
	if err != nil {
		return nil, 0, err
	}
	defer conn.Close()

	client := pb.NewAppServiceClient(conn)

	bootstrapErr := bootstrapConnect(ctx, client, cfg)
	if bootstrapErr != nil {
		logger.Warnf("bootstrap connect failed: %v", bootstrapErr)
	}

	resp, err := client.GetInternalPeerInfo(ctx, &pb.GetInternalPeerInfoRequest{})
	if err != nil {
		return bootstrapTargets, int32(cfg.P2PPort), nil
	}

	ips := map[string]struct{}{}
	port := int32(cfg.P2PPort)
	for ip := range bootstrapTargets {
		ips[ip] = struct{}{}
	}

	for _, entry := range resp.GetEntries() {
		if entry == nil {
			continue
		}

		if len(cfg.AllowedPeerID) > 0 {
			if _, ok := cfg.AllowedPeerID[entry.PeerId]; !ok {
				continue
			}
		} else {
			if !contains(entry.SupportedServices, "miner") {
				continue
			}
			if entry.ConnectionState != "connected" {
				continue
			}
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

		addr := ap.Addr()
		if !addr.Is4() {
			continue
		}
		if p := int32(ap.Port()); p > 0 {
			port = p
		}
		ips[addr.String()] = struct{}{}
	}

	return ips, port, nil
}

// parseBootstrapTargets resolves bootstrap endpoints to IP addresses.
func parseBootstrapTargets(ctx context.Context, cfg Config) map[string]struct{} {
	res := map[string]struct{}{}

	endpoints := make([]string, 0, len(cfg.Bootstrap.Endpoints))
	endpoints = append(endpoints, cfg.Bootstrap.Endpoints...)
	if len(endpoints) == 0 {
		for _, ip := range cfg.OverrideIPs {
			ip = strings.TrimSpace(ip)
			if ip == "" {
				continue
			}
			endpoints = append(endpoints, net.JoinHostPort(ip, strconv.Itoa(int(cfg.P2PPort))))
		}
	}

	for _, token := range endpoints {
		host, port, err := splitHostPortOrDefault(token, int(cfg.P2PPort))
		if err != nil {
			continue
		}
		_ = port

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

	return res
}

// bootstrapConnect attempts to connect to bootstrap endpoints via the app service.
func bootstrapConnect(ctx context.Context, client pb.AppServiceClient, cfg Config) error {
	endpoints := make([]string, 0, len(cfg.Bootstrap.Endpoints))
	endpoints = append(endpoints, cfg.Bootstrap.Endpoints...)
	if len(endpoints) == 0 {
		for _, ip := range cfg.OverrideIPs {
			ip = strings.TrimSpace(ip)
			if ip == "" {
				continue
			}
			endpoints = append(endpoints, net.JoinHostPort(ip, strconv.Itoa(int(cfg.P2PPort))))
		}
	}

	var lastErr error
	for _, token := range endpoints {
		host, port, err := splitHostPortOrDefault(token, int(cfg.P2PPort))
		if err != nil {
			lastErr = err
			continue
		}

		ips := []netip.Addr{}
		if parsed, err := netip.ParseAddr(host); err == nil {
			ips = append(ips, parsed)
		} else {
			resolved, err := net.DefaultResolver.LookupIPAddr(ctx, host)
			if err != nil {
				lastErr = err
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
				ips = append(ips, addr)
			}
		}

		for _, ip := range ips {
			if !ip.Is4() && !ip.Is6() {
				continue
			}
			resp, err := client.ConnectTo(ctx, &pb.ConnectToRequest{IpAddress: ip.AsSlice(), Port: uint32(port)})
			if err != nil {
				lastErr = err
				continue
			}
			if resp != nil && resp.Success {
				return nil
			}
			if resp != nil && !resp.Success {
				lastErr = fmt.Errorf("connect_to failed: %s", strings.TrimSpace(resp.ErrorMessage))
			}
		}
	}

	return lastErr
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

// dialAppGRPC establishes a gRPC connection to the app service.
func dialAppGRPC(ctx context.Context, addr string, useTLS bool) (*grpc.ClientConn, error) {
	if useTLS {
		creds := credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12})
		return grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(creds))
	}

	return grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
}

// contains checks whether needle exists in items.
func contains(items []string, needle string) bool {
	for _, item := range items {
		if item == needle {
			return true
		}
	}
	return false
}
