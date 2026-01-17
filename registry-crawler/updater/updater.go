package updater

import (
	"context"
	"sort"
	"strings"
	"time"

	"s3b/vsp-blockchain/registry-crawler/common"
	"s3b/vsp-blockchain/registry-crawler/discovery"
	"s3b/vsp-blockchain/registry-crawler/internal/pb"
	"s3b/vsp-blockchain/registry-crawler/peer"

	"bjoernblessin.de/go-utils/util/logger"
)

// peerManager is the global peer manager instance.
var peerManager *peer.PeerManager

// initPeerManager initializes the global peer manager with the configured TTL.
func initPeerManager(cfg common.Config) {
	if peerManager == nil {
		peerManager = peer.NewPeerManager(cfg.PeerKnownTTL)
	}
}

// RunPeerDiscoveryLoop runs the peer discovery loop in the background.
// Every PeerDiscoveryInterval, it attempts to connect to one new peer.
func RunPeerDiscoveryLoop(cfg common.Config) {
	logger.Infof("starting peer discovery loop interval: %s, known TTL: %s, registry subset size: %d",
		cfg.PeerDiscoveryInterval, cfg.PeerKnownTTL, cfg.PeerRegistrySubsetSize)

	initPeerManager(cfg)

	ticker := time.NewTicker(cfg.PeerDiscoveryInterval)
	defer ticker.Stop()

	for {
		logger.Tracef("=== peer discovery tick ===")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		discoverOnePeer(ctx, cfg)
		cancel()

		<-ticker.C
	}
}

// discoverOnePeer attempts to discover and verify one peer.
// It first tries to connect to StateNew peers, then re-verifies expired known peers.
func discoverOnePeer(ctx context.Context, cfg common.Config) {
	initPeerManager(cfg)

	peerManager.CleanupExpired()

	bootstrapPeers, port := discovery.ResolveBootstrapEndpoints(ctx, cfg)
	if len(bootstrapPeers) > 0 {
		added := peerManager.AddPeers(bootstrapPeers, port)
		if added > 0 {
			logger.Infof("discovered %d new bootstrap peers", added)
		}
	}

	networkPeers, port, err := discovery.FetchNetworkPeers(ctx, cfg)
	if err != nil {
		logger.Warnf("failed to fetch network peers: %v", err)
	}

	if len(networkPeers) > 0 {
		added := peerManager.AddPeers(networkPeers, port)
		if added > 0 {
			logger.Infof("discovered %d new network peers", added)
		}
	}

	peer := peerManager.GetNextUnverifiedPeer()
	if peer == nil {
		peer = peerManager.GetExpiredKnownPeer()
	}
	if peer == nil {
		logger.Tracef("no peers to (re-)verify this cycle")
		return
	}

	logger.Tracef("attempting to verify peer %s:%d", peer.IP, peer.Port)

	success := verifyPeer(ctx, cfg, peer.IP, peer.Port)
	if success {
		peerManager.MarkConnected(peer.IP)
		logger.Debugf("peer verified and marked as known: %s", peer.IP)
	} else {
		peerManager.MarkFailed(peer.IP)
		logger.Warnf("peer verification failed, removed: %s", peer.IP)
	}

	total, newCount, connecting, known := peerManager.Stats()
	logger.Infof("peer stats: total=%d new=%d connecting=%d known=%d", total, newCount, connecting, known)
}

// verifyPeer attempts to connect to a peer to verify it is reachable.
// Returns true if the peer responds correctly or is already connected.
func verifyPeer(ctx context.Context, cfg common.Config, ip string, port int32) bool {
	conn, err := discovery.DialAppGRPC(ctx, cfg.AppAddr)
	if err != nil {
		logger.Warnf("failed to dial app service: %v", err)
		return false
	}

	client := pb.NewAppServiceClient(conn)

	success, err := discovery.ConnectToPeer(ctx, client, ip, port)
	if err != nil {
		logger.Warnf("peer verification failed for %s:%d: %v", ip, port, err)
	}
	return success
}

// RunSeedUpdaterLoop periodically fetches seed targets and writes the DNS hosts file.
// This loop runs indefinitely, updating the hosts file at the configured interval.
func RunSeedUpdaterLoop(cfg common.Config) {
	logger.Infof("starting seed updater loop with interval=%s", cfg.UpdateEvery)
	initPeerManager(cfg)

	ticker := time.NewTicker(cfg.UpdateEvery)
	defer ticker.Stop()

	for {

		logger.Tracef("=== seed updater tick ===")
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		err := updateSeedHostsOnce(ctx, cfg)
		cancel()
		if err != nil {
			logger.Warnf("seed update failed: %v", err)
		}

		<-ticker.C
	}
}

// updateSeedHostsOnce fetches seed targets from known peers and writes the DNS hosts file once.
func updateSeedHostsOnce(ctx context.Context, cfg common.Config) error {
	var seedIPs map[string]struct{}
	seedPort := int32(cfg.AcceptedP2PPort)

	seedIPs = peerManager.GetRandomKnownPeerIPsFilteredByPort(seedPort, cfg.PeerRegistrySubsetSize)
	logger.Debugf("selected %d random known peers for seed targets (requested %d)", len(seedIPs), cfg.PeerRegistrySubsetSize)
	if len(seedIPs) == 0 {
		logger.Debugf("no known peers, falling back to bootstrap targets")
		bootstrapIPs, _ := discovery.ResolveBootstrapEndpoints(ctx, cfg)
		for ip := range bootstrapIPs {
			seedIPs[ip] = struct{}{}
		}
	}

	addresses := make([]string, 0, len(seedIPs))
	for ip := range seedIPs {
		addresses = append(addresses, ip)
	}
	sort.Strings(addresses)

	source := determineSource(cfg, len(peerManager.GetKnownPeers()))

	if strings.TrimSpace(cfg.SeedHostsFile) == "" {
		return nil
	}

	dnsAddresses := addresses
	if cfg.SeedDNSDebug.Enabled {
		dnsAddresses = generateRandomIPv4s(cfg.SeedDNSDebug.CIDR, cfg.SeedDNSDebug.Count)
		source = "debug-random"
	}

	hostsBody, err := buildSeedHostsFile(cfg.SeedName, cfg.SeedNamespace, cfg.SeedDNSZone, dnsAddresses, source)
	if err != nil {
		return err
	}

	if err := writeFileAtomically(cfg.SeedHostsFile, []byte(hostsBody)); err != nil {
		return err
	}

	logger.Infof("seed targets updated: port=%d source=%s addrs=%s file=%s", seedPort, source, strings.Join(addresses, ","), cfg.SeedHostsFile)

	return nil
}

// determineSource returns a string describing the source of seed targets.
func determineSource(cfg common.Config, knownPeerCount int) string {
	if cfg.SeedDNSDebug.Enabled {
		return "debug-random"
	}

	if knownPeerCount > 0 {
		return "known-peers"
	}

	source := "bootstrap"
	if len(cfg.Bootstrap.Endpoints) > 0 {
		source = source + "+endpoints"
	}
	return source
}
