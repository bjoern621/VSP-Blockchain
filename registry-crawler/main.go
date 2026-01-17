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
	"s3b/vsp-blockchain/registry-crawler/common"
	"s3b/vsp-blockchain/registry-crawler/updater"

	"bjoernblessin.de/go-utils/util/logger"
)

func main() {
	logger.Infof("[registry_crawler] Running registry crawler...")

	cfg := common.CurrentConfig()

	go updater.RunPeerDiscoveryLoop(cfg)
	go updater.RunSeedUpdaterLoop(cfg)

	select {}
}
