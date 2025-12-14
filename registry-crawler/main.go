// Package main implements the registry-crawler service.
//
// The registry-crawler periodically queries the P2P network for active miner nodes
// and maintains a DNS hosts file consumed by CoreDNS (reloaded every 5 seconds).
// This enables seed DNS resolution for new nodes joining the network.
//
// Architecture:
//
//	registry-crawler → queries → minimal_node (gRPC)
//	                 → writes  → /seed/seed.hosts (shared volume)
//	                                    ↓
//	seed-dns (CoreDNS) ← reads (5s) ← /seed/seed.hosts
package main

import (
	"bjoernblessin.de/go-utils/util/logger"
)

func main() {
	logger.Infof("Running registry crawler...")

	cfg := CurrentConfig()

	logger.Infof("peer discovery interval: %s, known TTL: %s, registry subset size: %d",
		cfg.PeerDiscoveryInterval, cfg.PeerKnownTTL, cfg.PeerRegistrySubsetSize)

	go runPeerDiscoveryLoop(cfg)
	go runSeedUpdaterLoop(cfg)

	select {}
}
