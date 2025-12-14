package main

import (
	"context"
	"sort"
	"strings"
	"time"

	"bjoernblessin.de/go-utils/util/logger"
)

// runSeedUpdaterLoop periodically fetches seed targets and writes the DNS hosts file.
// This loop runs indefinitely, updating the hosts file at the configured interval.
func runSeedUpdaterLoop(cfg Config) {
	ticker := time.NewTicker(cfg.UpdateEvery)
	defer ticker.Stop()

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		err := updateSeedHostsOnce(ctx, cfg)
		cancel()
		if err != nil {
			logger.Warnf("seed update failed: %v", err)
		}

		<-ticker.C
	}
}

// updateSeedHostsOnce fetches seed targets and writes the DNS hosts file once.
func updateSeedHostsOnce(ctx context.Context, cfg Config) error {
	seedIPs, seedPort, err := fetchSeedTargets(ctx, cfg)
	if err != nil {
		return err
	}

	addresses := make([]string, 0, len(seedIPs))
	for ip := range seedIPs {
		addresses = append(addresses, ip)
	}
	sort.Strings(addresses)

	source := determineSource(cfg)
	logger.Infof("seed targets: port=%d source=%s addrs=%s", seedPort, source, strings.Join(addresses, ","))

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

	logger.Infof("seed hosts written: %s", cfg.SeedHostsFile)
	return nil
}

// determineSource returns a string describing the source of seed targets.
func determineSource(cfg Config) string {
	if cfg.SeedDNSDebug.Enabled {
		return "debug-random"
	}

	source := "registry"
	if len(cfg.OverrideIPs) > 0 {
		source = "override"
	}
	if len(cfg.Bootstrap.Endpoints) > 0 {
		source = source + "+bootstrap"
	}
	return source
}
