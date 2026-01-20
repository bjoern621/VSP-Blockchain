package common

import (
	"math"
	"net/netip"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"

	"bjoernblessin.de/go-utils/util/assert"
	"bjoernblessin.de/go-utils/util/env"
	"bjoernblessin.de/go-utils/util/logger"
)

const (
	appPortEnvVar              = "APP_PORT"
	p2pPortEnvVar              = "P2P_PORT"
	appListenAddrEnvVar        = "APP_LISTEN_ADDR" // a IP address the app server binds to, can be 127.0.0.1
	p2pListenAddrEnvVar        = "P2P_LISTEN_ADDR" // a routable IP address the P2P server binds to
	additionalServicesEnvVar   = "ADDITIONAL_SERVICES"
	registrySeedHostnameEnvVar = "REGISTRY_SEED_HOSTNAME" // default: "miner-seed.seed.local."
)

var (
	appPort              atomic.Uint32
	p2pPort              atomic.Uint32
	appListenAddr        atomic.Value // string
	p2pListenAddr        atomic.Value // string
	additionalServices   atomic.Value // []string
	initialized          atomic.Bool
	registrySeedHostname atomic.Value // string
)

// Init reads all environment variables at startup.
// Read values are stored in package-level variables for easy access.
// This function should be called once at application startup.
func Init() {
	if initialized.Swap(true) {
		return
	}

	p2pPort.Store(uint32(readP2PPort()))
	p2pListenAddr.Store(readListenAddr(p2pListenAddrEnvVar))

	services := readAdditionalServices()
	validateAddionalServices(services)
	additionalServices.Store(services)

	if AppEnabled() {
		appPort.Store(uint32(readAppPort()))
		appListenAddr.Store(readListenAddr(appListenAddrEnvVar))
	}

	registrySeedHostname.Store(readRegistrySeedHostname())
}

func readAdditionalServices() []string {
	raw := strings.TrimSpace(os.Getenv(additionalServicesEnvVar))
	if raw == "" {
		return []string{}
	}

	parts := strings.Split(raw, ",")
	services := make([]string, 0, len(parts))
	for _, part := range parts {
		svc := strings.TrimSpace(part)

		switch svc {
		case "blockchain_full", "wallet", "miner", "app":
			services = append(services, svc)
		default:
			logger.Errorf("unknown service in %s: %s", additionalServicesEnvVar, svc)
		}
	}

	return services
}

func readRegistrySeedHostname() string {
	raw, found := env.ReadOptionalEnv(registrySeedHostnameEnvVar)
	if !found {
		return "miner-seed.seed.local."
	}

	raw = strings.TrimSpace(raw)
	if raw == "" {
		logger.Errorf("%s cannot be empty", registrySeedHostnameEnvVar)
	}

	return raw
}

func validateAddionalServices(services []string) {
	seen := make(map[string]struct{})
	for _, svc := range services {
		if _, exists := seen[svc]; exists {
			logger.Errorf("duplicate service in %s: %s", additionalServicesEnvVar, svc)
		}
		seen[svc] = struct{}{}
	}

	_, hasWallet := seen["wallet"]
	_, hasMiner := seen["miner"]
	_, hasBlockchainFull := seen["blockchain_full"]

	needsBlockchain := hasWallet || hasMiner
	if needsBlockchain && !hasBlockchainFull {
		logger.Errorf("wallet or miner service requires blockchain_full to be enabled")
	}
}

func getAdditionalServices() []string {
	assertInitialized()
	return slices.Clone(additionalServices.Load().([]string))
}

func AppPort() uint16 {
	assertInitialized()
	return uint16(appPort.Load())
}

func P2PPort() uint16 {
	assertInitialized()
	return uint16(p2pPort.Load())
}

func AppListenAddr() string {
	assertInitialized()
	return appListenAddr.Load().(string)
}

func P2PListenAddr() string {
	assertInitialized()
	return p2pListenAddr.Load().(string)
}

func RegistrySeedHostnameEnv() string {
	assertInitialized()
	return registrySeedHostname.Load().(string)
}

func assertInitialized() {
	assert.Assert(initialized.Load(), "common.Init() must be called before accessing environment variables")
}

// readAppPort reads the application port used by the app endpoint from the environment variable appPortEnvVar.
// Environment variable is optional. If
//   - 0 is provided, 0 is returned.
//   - no value is provided, the default port is used.
//   - invalid value is provided, the application logs an error and exits.
func readAppPort() uint16 {
	return readUint16EnvOrDefault(appPortEnvVar, defaultAppPort)
}

// readP2PPort is similar to readAppPort but for the P2P port.
func readP2PPort() uint16 {
	return readUint16EnvOrDefault(p2pPortEnvVar, DefaultP2PPort)
}

func readUint16EnvOrDefault(key string, fallback uint16) uint16 {
	portStr, found := env.ReadOptionalEnv(key)

	if !found {
		return fallback
	}

	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		logger.Errorf("invalid %s value: %s, must be between 0 and 65535", key, portStr)
	}

	assert.Assert(port <= math.MaxUint16, "port value %d out of range", port)

	return uint16(port)
}

func readListenAddr(key string) string {
	raw := env.ReadNonEmptyRequiredEnv(key)

	if _, err := netip.ParseAddr(raw); err != nil {
		logger.Warnf("invalid %s value: %s", key, raw)
	}

	return raw
}

// SetAppPort sets the application port to the given value.
// Is needed for example for dynamic port assignment.
func SetAppPort(port uint16) {
	appPort.Store(uint32(port))
}

func SetP2PPort(port uint16) {
	p2pPort.Store(uint32(port))
}
