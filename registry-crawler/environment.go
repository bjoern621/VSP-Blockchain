package main

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

const (
	appGrpcAddrPortEnvVar        = "APP_GRPC_ADDR_PORT"
	p2pPortEnvVar                = "P2P_PORT"
	k8sDisabledEnvVar            = "K8S_DISABLED"
	seedNamespaceEnvVar          = "SEED_NAMESPACE"
	seedEndpointsNameEnvVar      = "SEED_ENDPOINTS_NAME"
	seedDNSConfigMapEnvVar       = "SEED_DNS_CONFIGMAP_NAME"
	seedDNSHostsKeyEnvVar        = "SEED_DNS_HOSTS_KEY"
	seedDNSZoneEnvVar            = "SEED_DNS_ZONE"
	seedUpdateIntervalEnvVar     = "SEED_UPDATE_INTERVAL"
	seedAllowedPeerIDsEnvVar     = "SEED_ALLOWED_PEER_IDS"
	seedOverrideIPsEnvVar        = "SEED_OVERRIDE_IPS"
	seedBootstrapEndpointsEnvVar = "SEED_BOOTSTRAP_ENDPOINTS"
	appGrpcTLSEnvVar             = "APP_GRPC_TLS"
	seedDNSDebugRandomIPsEnvVar  = "SEED_DNS_DEBUG_RANDOM_IPS"
	seedDNSDebugCountEnvVar      = "SEED_DNS_DEBUG_COUNT"
	seedDNSDebugCIDREnvVar       = "SEED_DNS_DEBUG_CIDR"
)

var (
	appGrpcAddr atomic.Value // string
	p2pPort     atomic.Uint32

	k8sDisabled atomic.Bool
	appGrpcTLS  atomic.Bool

	seedNamespace atomic.Value // string
	seedName      atomic.Value // string
	seedDNSConfig atomic.Value // string
	seedDNSKey    atomic.Value // string
	seedDNSZone   atomic.Value // string

	seedUpdateIntervalNs atomic.Int64

	seedAllowedPeerIDs atomic.Value // map[string]struct{}
	seedOverrideIPs    atomic.Value // []string
	seedBootstrapEPs   atomic.Value // []string
	seedDNSDebug       atomic.Value // dnsDebugConfig
)

func init() {
	cfg := readAndValidateEnvironment()

	appGrpcAddr.Store(cfg.appAddr)
	p2pPort.Store(uint32(cfg.p2pPort))

	k8sDisabled.Store(cfg.k8sDisabled)
	appGrpcTLS.Store(cfg.appGrpcTLS)

	seedNamespace.Store(cfg.seedNamespace)
	seedName.Store(cfg.seedName)
	seedDNSConfig.Store(cfg.seedDNSConfig)
	seedDNSKey.Store(cfg.seedDNSKey)
	seedDNSZone.Store(cfg.seedDNSZone)

	seedUpdateIntervalNs.Store(cfg.seedUpdateEvery.Nanoseconds())

	seedAllowedPeerIDs.Store(cfg.allowedPeerIDs)
	seedOverrideIPs.Store(cfg.overrideIPs)
	seedBootstrapEPs.Store(cfg.bootstrapEndpoints)
	seedDNSDebug.Store(cfg.seedDNSDebug)
}

type envSnapshot struct {
	appAddr string
	p2pPort uint16

	k8sDisabled bool
	appGrpcTLS  bool

	seedNamespace string
	seedName      string
	seedDNSConfig string
	seedDNSKey    string
	seedDNSZone   string

	seedUpdateEvery time.Duration

	allowedPeerIDs     map[string]struct{}
	overrideIPs        []string
	bootstrapEndpoints []string
	seedDNSDebug       dnsDebugConfig
}

func readAndValidateEnvironment() envSnapshot {
	appAddr := env.ReadNonEmptyRequiredEnv(appGrpcAddrPortEnvVar)
	p2p := mustReadRequiredPort(p2pPortEnvVar)
	interval := mustReadRequiredDuration(seedUpdateIntervalEnvVar)

	k8sOff := readOptionalBool(k8sDisabledEnvVar)
	tlsEnabled := readOptionalBool(appGrpcTLSEnvVar)

	seedNS := defaultSeedNamespace
	seedName := defaultSeedEndpointsName
	seedDNSConfig := defaultSeedDNSConfigMap
	seedDNSKey := defaultSeedDNSHostsKey
	seedDNSZone := defaultSeedDNSZone
	if !k8sOff {
		seedNS = env.ReadNonEmptyRequiredEnv(seedNamespaceEnvVar)
		seedName = env.ReadNonEmptyRequiredEnv(seedEndpointsNameEnvVar)
		seedDNSConfig = env.ReadNonEmptyRequiredEnv(seedDNSConfigMapEnvVar)
		seedDNSKey = env.ReadNonEmptyRequiredEnv(seedDNSHostsKeyEnvVar)
		seedDNSZone = env.ReadNonEmptyRequiredEnv(seedDNSZoneEnvVar)
	}

	allowedPeerIDs := map[string]struct{}{}
	if raw, ok := env.ReadOptionalEnv(seedAllowedPeerIDsEnvVar); ok {
		raw = strings.TrimSpace(raw)
		for token := range strings.SplitSeq(raw, ",") {
			peerID := strings.TrimSpace(token)
			if peerID == "" {
				continue
			}
			allowedPeerIDs[peerID] = struct{}{}
		}
	}

	overrideIPs := []string{}
	if raw, ok := env.ReadOptionalEnv(seedOverrideIPsEnvVar); ok {
		raw = strings.TrimSpace(raw)
		for token := range strings.SplitSeq(raw, ",") {
			ipString := strings.TrimSpace(token)
			if ipString == "" {
				continue
			}
			overrideIPs = append(overrideIPs, ipString)
		}
	}

	bootstrapEPs := []string{}
	if raw, ok := env.ReadOptionalEnv(seedBootstrapEndpointsEnvVar); ok {
		raw = strings.TrimSpace(raw)
		for token := range strings.SplitSeq(raw, ",") {
			ep := strings.TrimSpace(token)
			if ep == "" {
				continue
			}
			bootstrapEPs = append(bootstrapEPs, ep)
		}
	}

	dnsDebug := readDNSDebugConfigOrDie()

	return envSnapshot{
		appAddr:            appAddr,
		p2pPort:            p2p,
		k8sDisabled:        k8sOff,
		appGrpcTLS:         tlsEnabled,
		seedNamespace:      seedNS,
		seedName:           seedName,
		seedDNSConfig:      seedDNSConfig,
		seedDNSKey:         seedDNSKey,
		seedDNSZone:        seedDNSZone,
		seedUpdateEvery:    interval,
		allowedPeerIDs:     allowedPeerIDs,
		overrideIPs:        overrideIPs,
		bootstrapEndpoints: bootstrapEPs,
		seedDNSDebug:       dnsDebug,
	}
}

// mustReadRequiredPort reads an environment variable as uint16 port.
func mustReadRequiredPort(key string) uint16 {
	raw := env.ReadNonEmptyRequiredEnv(key)
	v, err := strconv.ParseUint(raw, 10, 16)
	if err != nil {
		logger.Errorf("%s: %v", key, err)
	}
	if v < 1 || v > math.MaxUint16 {
		logger.Errorf("%s out of range: %d", key, v)
	}
	return uint16(v)
}

// mustReadRequiredDuration reads an environment variable as time.Duration.
func mustReadRequiredDuration(key string) time.Duration {
	raw := env.ReadNonEmptyRequiredEnv(key)
	d, err := time.ParseDuration(raw)
	if err != nil {
		logger.Errorf("%s: %v", key, err)
	}
	return d
}

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

func readDNSDebugConfigOrDie() dnsDebugConfig {
	enabled := readOptionalBool(seedDNSDebugRandomIPsEnvVar)

	count := 2
	if raw, ok := env.ReadOptionalEnv(seedDNSDebugCountEnvVar); ok {
		raw = strings.TrimSpace(raw)
		v, err := strconv.Atoi(raw)
		if err != nil {
			logger.Errorf("%s: %v", seedDNSDebugCountEnvVar, err)
			panic(fmt.Sprintf("%s: %v", seedDNSDebugCountEnvVar, err))
		}
		count = v
	}
	if count < 1 {
		count = 1
	}
	if count > 20 {
		count = 20
	}

	cidr := "203.0.113.0/24" // TEST-NET-3
	if raw, ok := env.ReadOptionalEnv(seedDNSDebugCIDREnvVar); ok {
		raw = strings.TrimSpace(raw)
		if raw != "" {
			cidr = raw
		}
	}
	prefix, err := netip.ParsePrefix(cidr)
	if err != nil {
		logger.Errorf("%s: %v", seedDNSDebugCIDREnvVar, err)
		panic(fmt.Sprintf("%s: %v", seedDNSDebugCIDREnvVar, err))
	}
	if !prefix.Addr().Is4() {
		err := fmt.Sprintf("%s must be IPv4, got %s", seedDNSDebugCIDREnvVar, prefix.String())
		logger.Errorf("%s", err)
		panic(err)
	}

	return dnsDebugConfig{enabled: enabled, count: count, cidr: prefix}
}

func AppGRPCAddr() string {
	return appGrpcAddr.Load().(string)
}

func P2PPort() uint16 {
	return uint16(p2pPort.Load())
}

func K8SDisabled() bool {
	return k8sDisabled.Load()
}

func AppGRPCTLS() bool {
	return appGrpcTLS.Load()
}

func SeedNamespace() string {
	return seedNamespace.Load().(string)
}

func SeedName() string {
	return seedName.Load().(string)
}

func SeedDNSConfigMapName() string {
	return seedDNSConfig.Load().(string)
}

func SeedDNSHostsKey() string {
	return seedDNSKey.Load().(string)
}

func SeedDNSZone() string {
	return seedDNSZone.Load().(string)
}

func SeedUpdateInterval() time.Duration {
	return time.Duration(seedUpdateIntervalNs.Load())
}

func SeedAllowedPeerIDs() map[string]struct{} {
	return seedAllowedPeerIDs.Load().(map[string]struct{})
}

func SeedOverrideIPs() []string {
	return seedOverrideIPs.Load().([]string)
}

func SeedBootstrapEndpoints() []string {
	return seedBootstrapEPs.Load().([]string)
}

func SeedDNSDebug() dnsDebugConfig {
	return seedDNSDebug.Load().(dnsDebugConfig)
}

func CurrentConfig() config {
	return config{
		appAddr:       AppGRPCAddr(),
		p2pPort:       P2PPort(),
		seedNamespace: SeedNamespace(),
		seedName:      SeedName(),
		seedDNSConfig: SeedDNSConfigMapName(),
		seedDNSKey:    SeedDNSHostsKey(),
		seedDNSZone:   SeedDNSZone(),
		seedDNSDebug:  SeedDNSDebug(),
		bootstrap:     bootstrapConfig{endpoints: SeedBootstrapEndpoints()},
		updateEvery:   SeedUpdateInterval(),
		allowedPeerID: SeedAllowedPeerIDs(),
		overrideIPs:   SeedOverrideIPs(),
		useTLS:        AppGRPCTLS(),
	}
}
