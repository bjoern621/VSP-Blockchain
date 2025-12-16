package discovery

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"slices"
	"strconv"
	"strings"

	"s3b/vsp-blockchain/registry-crawler/common"
	"s3b/vsp-blockchain/registry-crawler/internal/pb"

	"bjoernblessin.de/go-utils/util/logger"
)

// ResolveBootstrapEndpoints resolves bootstrap endpoints to IP addresses.
// Returns a set of IP addresses and the P2P port.
func ResolveBootstrapEndpoints(ctx context.Context, cfg common.Config) (map[string]struct{}, int32) {
	res := map[string]struct{}{}
	acceptedPort := cfg.AcceptedP2PPort

	for _, endpoint := range cfg.Bootstrap.Endpoints {
		resolveEndpointToIPv4Set(ctx, endpoint, acceptedPort, res)
	}

	return res, int32(acceptedPort)
}

// resolveEndpointToIPv4Set resolves a single endpoint to IPv4 addresses and adds them to the result set.
func resolveEndpointToIPv4Set(ctx context.Context, endpoint string, acceptedPort uint16, result map[string]struct{}) {
	host, port, err := splitHostPortOrDefault(endpoint, int(acceptedPort))
	if err != nil || uint16(port) != acceptedPort {
		return
	}

	if tryAddDirectIPv4(host, result) {
		return
	}

	resolveDNSToIPv4Set(ctx, host, result)
}

// tryAddDirectIPv4 attempts to parse the host as an IPv4 address and add it to the result set.
// Returns true if the host was a valid IP address.
func tryAddDirectIPv4(host string, result map[string]struct{}) bool {
	parsed, err := netip.ParseAddr(host)
	if err != nil {
		return false
	}
	if parsed.Is4() {
		result[parsed.String()] = struct{}{}
	}
	return true
}

// resolveDNSToIPv4Set performs a DNS lookup and adds all resolved IPv4 addresses to the result set.
func resolveDNSToIPv4Set(ctx context.Context, host string, result map[string]struct{}) {
	resolved, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return
	}

	for _, r := range resolved {
		addIPv4ToSet(r.IP, result)
	}
}

// addIPv4ToSet adds an IP address to the result set if it is a valid IPv4 address.
func addIPv4ToSet(ip net.IP, result map[string]struct{}) {
	if ip == nil {
		return
	}
	addr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return
	}
	if addr.Is4() {
		result[addr.String()] = struct{}{}
	}
}

// FetchNetworkPeers queries the app service for connected peers from the P2P network.
// Returns a set of IP addresses and the P2P port.
// Filters out:
// - peers that do not support "blockchain_full" service
// - peers that do not use the accepted P2P port (standard port)
func FetchNetworkPeers(ctx context.Context, cfg common.Config) (map[string]struct{}, int32, error) {
	conn, err := DialAppGRPC(ctx, cfg.AppAddr)
	if err != nil {
		return nil, 0, err
	}

	client := pb.NewAppServiceClient(conn)

	resp, err := client.GetInternalPeerInfo(ctx, &pb.GetInternalPeerInfoRequest{})
	if err != nil {
		return nil, int32(cfg.AcceptedP2PPort), nil
	}

	logger.Tracef("fetched %d peers from app service total", len(resp.GetEntries()))

	ips := map[string]struct{}{}
	acceptedPort := uint16(cfg.AcceptedP2PPort)

	for _, entry := range resp.GetEntries() {
		if entry == nil {
			continue
		}

		if !slices.Contains(entry.SupportedServices, "blockchain_full") {
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

		ips[addr.String()] = struct{}{}
	}

	logger.Tracef("fetched %d usable network peers from app service: %v", len(ips), ips)

	return ips, int32(cfg.AcceptedP2PPort), nil
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
