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
//	seed-dns (CoreDNS) ← reads ← /seed/seed.hosts
package main

import (
	"bjoernblessin.de/go-utils/util/logger"
)

func main() {
	logger.Infof("Running registry crawler...")

	cfg := CurrentConfig()

	go runSeedUpdaterLoop(cfg)

	select {}
}
