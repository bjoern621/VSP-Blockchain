package main

import (
	"net/netip"
	"time"
)

// Config holds the runtime configuration for the registry-crawler.
type Config struct {
	// AppAddr is the gRPC address of the app service to query for peer information.
	AppAddr string
	// AcceptedP2PPort is the default P2P port used by nodes. Only peers using this port are accepted.
	AcceptedP2PPort uint16
	// SeedHostsFile is the path to write the DNS hosts file.
	SeedHostsFile string
	// SeedNamespace is the namespace identifier for DNS records.
	SeedNamespace string
	// SeedName is the service name used in DNS records.
	SeedName string
	// SeedDNSZone is the DNS zone suffix for seed records.
	SeedDNSZone string
	// SeedDNSDebug contains debug configuration for generating random IPs.
	SeedDNSDebug DNSDebugConfig
	// Bootstrap contains bootstrap peer endpoints.
	Bootstrap BootstrapConfig
	// UpdateEvery is the interval between seed registry updates.
	UpdateEvery time.Duration
	// PeerDiscoveryInterval is the interval between peer discovery attempts.
	PeerDiscoveryInterval time.Duration
	// PeerKnownTTL is the TTL for known peers before re-verification.
	PeerKnownTTL time.Duration
	// PeerRegistrySubsetSize is the number of random known peers to use for registry updates.
	PeerRegistrySubsetSize int
}

// BootstrapConfig holds bootstrap peer configuration.
type BootstrapConfig struct {
	// Endpoints are the initial peer addresses to connect to.
	Endpoints []string
}

// DNSDebugConfig controls debug IP generation for testing.
type DNSDebugConfig struct {
	// Enabled activates random IP generation instead of real peer discovery.
	Enabled bool
	// Count is the number of random IPs to generate.
	Count int
	// CIDR is the IP range for random generation.
	CIDR netip.Prefix
}
