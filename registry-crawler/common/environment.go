package common

import (
	"fmt"
	"math"
	"net/netip"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"bjoernblessin.de/go-utils/util/env"
	"bjoernblessin.de/go-utils/util/logger"
)

// Environment variable names for configuration.
const (
	// Required. string, gRPC address of the app service (host:port).
	appGrpcAddrPortEnvVar = "APP_GRPC_ADDR_PORT"
	// Required. uint16, P2P port used by miner nodes. Range: 1-65535.
	acceptedP2pPortEnvVar = "ACCEPTED_P2P_PORT"

	// Optional. string, path to write the DNS hosts file. Empty disables file output.
	seedHostsFileEnvVar = "SEED_HOSTS_FILE"

	// Optional. string, namespace identifier for DNS records. Default: "vsp-blockchain".
	seedNamespaceEnvVar = "SEED_NAMESPACE"
	// Optional. string, service name used in DNS records. Default: "miner-seed".
	seedEndpointsNameEnvVar = "SEED_ENDPOINTS_NAME"
	// Optional. string, DNS zone suffix for seed records. Default: "seed.local".
	seedDNSZoneEnvVar = "SEED_DNS_ZONE"

	// Required. duration, interval between seed updates. Format: Go duration (e.g. "30s", "1m").
	seedUpdateIntervalEnvVar = "SEED_UPDATE_INTERVAL"
	// Optional. duration, interval between peer discovery attempts. Default: 30s. Should be lower than SEED_UPDATE_INTERVAL.
	peerDiscoveryIntervalEnvVar = "PEER_DISCOVERY_INTERVAL"
	// Optional. duration, TTL for known peers before re-verification. Default: 15m.
	peerKnownTTLEnvVar = "PEER_KNOWN_TTL"
	// Optional. int, number of random known peers to use for registry updates. Default: 5.
	peerRegistrySubsetSizeEnvVar = "PEER_REGISTRY_SUBSET_SIZE"
	// Optional. string, comma-separated list of bootstrap endpoints (host:port). This allows the crawler to connect to initial peers.
	seedBootstrapEndpointsEnvVar = "SEED_BOOTSTRAP_ENDPOINTS"

	// Optional. boolean, enables random IP generation for testing. Values: "true" or "false". Default: false.
	seedDNSDebugRandomIPsEnvVar = "SEED_DNS_DEBUG_RANDOM_IPS"
)

// Default configuration values.
const (
	defaultSeedNamespace          = "vsp-blockchain"
	defaultSeedEndpointsName      = "miner-seed"
	defaultSeedDNSZone            = "seed.local"
	defaultDebugCIDR              = "203.0.113.0/24" // TEST-NET-3
	defaultDebugRandomIPCount     = 6
	defaultPeerDiscoveryInterval  = 30 * time.Second
	defaultPeerKnownTTL           = 15 * time.Minute
	defaultPeerRegistrySubsetSize = 5
)

// Atomic configuration storage.
var (
	appGrpcAddr     atomic.Value // string
	acceptedP2pPort atomic.Uint32
	seedHostsFile   atomic.Value // string
	seedNamespace   atomic.Value // string
	seedName        atomic.Value // string
	seedDNSZone     atomic.Value // string

	seedUpdateIntervalNs    atomic.Int64
	peerDiscoveryIntervalNs atomic.Int64
	peerKnownTTLNs          atomic.Int64
	peerRegistrySubsetSize  atomic.Int32

	seedBootstrapEPs atomic.Value // []string
	seedDNSDebug     atomic.Value // DNSDebugConfig
)

func init() {
	cfg := readAndValidateEnvironment()

	appGrpcAddr.Store(cfg.appAddr)
	acceptedP2pPort.Store(uint32(cfg.acceptedP2pPort))
	seedHostsFile.Store(cfg.seedHostsFile)
	seedNamespace.Store(cfg.seedNamespace)
	seedName.Store(cfg.seedName)
	seedDNSZone.Store(cfg.seedDNSZone)
	seedUpdateIntervalNs.Store(cfg.seedUpdateEvery.Nanoseconds())
	peerDiscoveryIntervalNs.Store(cfg.peerDiscoveryInterval.Nanoseconds())
	peerKnownTTLNs.Store(cfg.peerKnownTTL.Nanoseconds())
	peerRegistrySubsetSize.Store(int32(cfg.peerRegistrySubsetSize))
	seedBootstrapEPs.Store(cfg.bootstrapEndpoints)
	seedDNSDebug.Store(cfg.seedDNSDebug)
}

type envSnapshot struct {
	appAddr                string
	acceptedP2pPort        uint16
	seedHostsFile          string
	seedNamespace          string
	seedName               string
	seedDNSZone            string
	seedUpdateEvery        time.Duration
	peerDiscoveryInterval  time.Duration
	peerKnownTTL           time.Duration
	peerRegistrySubsetSize int
	bootstrapEndpoints     []string
	seedDNSDebug           DNSDebugConfig
}

// readAndValidateEnvironment reads and validates all environment variables.
func readAndValidateEnvironment() envSnapshot {
	appAddr := env.ReadNonEmptyRequiredEnv(appGrpcAddrPortEnvVar)
	acceptedP2p := mustReadRequiredPort(acceptedP2pPortEnvVar)
	interval := mustReadRequiredDuration(seedUpdateIntervalEnvVar)

	peerDiscovery := readOptionalDurationWithDefault(peerDiscoveryIntervalEnvVar, defaultPeerDiscoveryInterval)
	peerTTL := readOptionalDurationWithDefault(peerKnownTTLEnvVar, defaultPeerKnownTTL)
	subsetSize := readOptionalIntWithDefault(peerRegistrySubsetSizeEnvVar, defaultPeerRegistrySubsetSize)

	seedHostsFileVal := ""
	if raw, ok := env.ReadOptionalEnv(seedHostsFileEnvVar); ok {
		seedHostsFileVal = strings.TrimSpace(raw)
	}

	seedNS := readOptionalStringWithDefault(seedNamespaceEnvVar, defaultSeedNamespace)
	seedNameVal := readOptionalStringWithDefault(seedEndpointsNameEnvVar, defaultSeedEndpointsName)
	seedDNSZoneVal := readOptionalStringWithDefault(seedDNSZoneEnvVar, defaultSeedDNSZone)

	bootstrapEPs := parseCommaSeparatedList(seedBootstrapEndpointsEnvVar)
	dnsDebug := DNSDebugConfig{
		Enabled: readOptionalBool(seedDNSDebugRandomIPsEnvVar),
		Count:   defaultDebugRandomIPCount,
		CIDR:    netip.MustParsePrefix(defaultDebugCIDR),
	}

	return envSnapshot{
		appAddr:                appAddr,
		acceptedP2pPort:        acceptedP2p,
		seedHostsFile:          seedHostsFileVal,
		seedNamespace:          seedNS,
		seedName:               seedNameVal,
		seedDNSZone:            seedDNSZoneVal,
		seedUpdateEvery:        interval,
		peerDiscoveryInterval:  peerDiscovery,
		peerKnownTTL:           peerTTL,
		peerRegistrySubsetSize: subsetSize,
		bootstrapEndpoints:     bootstrapEPs,
		seedDNSDebug:           dnsDebug,
	}
}

// mustReadRequiredPort reads an environment variable as uint16 port and panics on error.
func mustReadRequiredPort(key string) uint16 {
	raw := env.ReadNonEmptyRequiredEnv(key)
	v, err := strconv.ParseUint(raw, 10, 16)
	if err != nil {
		logger.Errorf("%s: %v", key, err)
		panic(fmt.Sprintf("%s: %v", key, err))
	}
	if v < 1 || v > math.MaxUint16 {
		logger.Errorf("%s out of range: %d", key, v)
		panic(fmt.Sprintf("%s out of range: %d", key, v))
	}
	return uint16(v)
}

// mustReadRequiredDuration reads an environment variable as time.Duration and panics on error.
func mustReadRequiredDuration(key string) time.Duration {
	raw := env.ReadNonEmptyRequiredEnv(key)
	d, err := time.ParseDuration(raw)
	if err != nil {
		logger.Errorf("%s: %v", key, err)
		panic(fmt.Sprintf("%s: %v", key, err))
	}
	return d
}

// readOptionalDurationWithDefault reads an optional duration environment variable with a default.
func readOptionalDurationWithDefault(key string, defaultValue time.Duration) time.Duration {
	raw, ok := env.ReadOptionalEnv(key)
	if !ok {
		return defaultValue
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return defaultValue
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		logger.Warnf("%s: invalid duration %q, using default %s: %v", key, raw, defaultValue, err)
		return defaultValue
	}
	return d
}

// readOptionalIntWithDefault reads an optional int environment variable with a default.
func readOptionalIntWithDefault(key string, defaultValue int) int {
	raw, ok := env.ReadOptionalEnv(key)
	if !ok {
		return defaultValue
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return defaultValue
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		logger.Warnf("%s: invalid int %q, using default %d: %v", key, raw, defaultValue, err)
		return defaultValue
	}
	return v
}

// readOptionalBool reads an optional boolean environment variable.
func readOptionalBool(key string) bool {
	raw, ok := env.ReadOptionalEnv(key)
	if !ok {
		return false
	}
	raw = strings.TrimSpace(raw)
	switch raw {
	case "true":
		return true
	case "false":
		return false
	default:
		logger.Errorf("%s: invalid boolean value: %s, must be 'true' or 'false'", key, raw)
		return false
	}
}

// readOptionalStringWithDefault reads an optional string environment variable with a default.
func readOptionalStringWithDefault(key, defaultValue string) string {
	if raw, ok := env.ReadOptionalEnv(key); ok {
		if trimmed := strings.TrimSpace(raw); trimmed != "" {
			return trimmed
		}
	}
	return defaultValue
}

// parseCommaSeparatedList parses a comma-separated environment variable into a slice.
func parseCommaSeparatedList(key string) []string {
	result := []string{}
	if raw, ok := env.ReadOptionalEnv(key); ok {
		raw = strings.TrimSpace(raw)
		for token := range strings.SplitSeq(raw, ",") {
			val := strings.TrimSpace(token)
			if val == "" {
				continue
			}
			result = append(result, val)
		}
	}
	return result
}

// CurrentConfig returns the current runtime configuration.
func CurrentConfig() Config {
	return Config{
		AppAddr:                appGrpcAddr.Load().(string),
		AcceptedP2PPort:        uint16(acceptedP2pPort.Load()),
		SeedHostsFile:          seedHostsFile.Load().(string),
		SeedNamespace:          seedNamespace.Load().(string),
		SeedName:               seedName.Load().(string),
		SeedDNSZone:            seedDNSZone.Load().(string),
		SeedDNSDebug:           seedDNSDebug.Load().(DNSDebugConfig),
		Bootstrap:              BootstrapConfig{Endpoints: seedBootstrapEPs.Load().([]string)},
		UpdateEvery:            time.Duration(seedUpdateIntervalNs.Load()),
		PeerDiscoveryInterval:  time.Duration(peerDiscoveryIntervalNs.Load()),
		PeerKnownTTL:           time.Duration(peerKnownTTLNs.Load()),
		PeerRegistrySubsetSize: int(peerRegistrySubsetSize.Load()),
	}
}
