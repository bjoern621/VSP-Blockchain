package discovery

import (
	"context"
	"net/netip"
	"s3b/vsp-blockchain/registry-crawler/common"
	"s3b/vsp-blockchain/registry-crawler/internal/pb"
	"slices"
	"strings"

	"bjoernblessin.de/go-utils/util/logger"
)

// FetchNetworkPeers queries the app service for connected peers from the P2P network.
// Returns a set of IP addresses and the P2P port.
// Filters out:
// - peers that do not support "blockchain_full" service
// - peers that do not use the accepted P2P port (standard port)
func FetchNetworkPeers(ctx context.Context, cfg common.Config) (map[string]struct{}, int32, error) {
	entries, err := fetchPeerEntries(ctx, cfg.AppAddr)
	if err != nil {
		return nil, 0, err
	}

	logger.Tracef("fetched %d peers from app service total", len(entries))

	acceptedPort := uint16(cfg.AcceptedP2PPort)
	ips := extractValidPeerIPs(entries, acceptedPort)

	logger.Tracef("fetched %d usable network peers from app service: %v", len(ips), ips)

	return ips, int32(cfg.AcceptedP2PPort), nil
}

// fetchPeerEntries establishes a gRPC connection and retrieves peer info entries.
func fetchPeerEntries(ctx context.Context, appAddr string) ([]*pb.InternalPeerInfoEntry, error) {
	conn, err := DialAppGRPC(ctx, appAddr)
	if err != nil {
		return nil, err
	}

	client := pb.NewAppServiceClient(conn)
	resp, err := client.GetInternalPeerInfo(ctx, &pb.GetInternalPeerInfoRequest{})
	if err != nil {
		return nil, nil
	}

	return resp.GetEntries(), nil
}

// extractValidPeerIPs filters peer entries and extracts IP addresses of valid peers.
// A peer is valid when it supports "blockchain_full" service and uses the accepted port.
func extractValidPeerIPs(entries []*pb.InternalPeerInfoEntry, acceptedPort uint16) map[string]struct{} {
	ips := map[string]struct{}{}

	for _, entry := range entries {
		if addr, ok := checkIfEntryHasValidIP(entry, acceptedPort); ok {
			ips[addr] = struct{}{}
		}
	}

	return ips
}

// checkIfEntryHasValidIP extracts the IP address from a single peer entry.
// Returns the IP string and true when the entry is valid, otherwise returns empty string and false.
func checkIfEntryHasValidIP(entry *pb.InternalPeerInfoEntry, acceptedPort uint16) (string, bool) {
	if entry == nil {
		return "", false
	}

	if !slices.Contains(entry.SupportedServices, "blockchain_full") {
		return "", false
	}

	endpoint, ok := getListeningEndpoint(entry)
	if !ok {
		return "", false
	}

	ap, err := netip.ParseAddrPort(strings.TrimSpace(endpoint))
	if err != nil {
		return "", false
	}

	if ap.Port() != acceptedPort {
		return "", false
	}

	return ap.Addr().String(), true
}

// getListeningEndpoint retrieves the listening endpoint from the entry infrastructure data.
func getListeningEndpoint(entry *pb.InternalPeerInfoEntry) (string, bool) {
	infra := entry.GetInfrastructureData()
	if infra == nil {
		return "", false
	}

	infraMap := infra.AsMap()
	endpoint, ok := infraMap["listeningEndpoint"].(string)
	return endpoint, ok
}
