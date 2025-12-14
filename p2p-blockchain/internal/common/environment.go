package common

import (
	"math"
	"net"
	"net/netip"
	"os"
	"strconv"
	"strings"
	"sync/atomic"

	"bjoernblessin.de/go-utils/util/assert"
	"bjoernblessin.de/go-utils/util/env"
	"bjoernblessin.de/go-utils/util/logger"
)

const (
	appPortEnvVar            = "APP_PORT"
	p2pPortEnvVar            = "P2P_PORT"
	appListenAddrEnvVar      = "APP_LISTEN_ADDR"
	p2pListenAddrEnvVar      = "P2P_LISTEN_ADDR"
	p2pAdvertiseIPEnvVar     = "P2P_ADVERTISE_IP"
	additionalServicesEnvVar = "ADDITIONAL_SERVICES"
)

var (
	appPort            atomic.Uint32
	p2pPort            atomic.Uint32
	appListenAddr      atomic.Value // string
	p2pListenAddr      atomic.Value // string
	additionalServices atomic.Value // []string
)

// init reads all environment variables at startup.
// Read values are stored in package-level variables for easy access.
func init() {
	appPort.Store(uint32(readAppPort()))
	p2pPort.Store(uint32(readP2PPort()))
	appListenAddr.Store(readListenAddrOrDefault(appListenAddrEnvVar, "127.0.0.1"))
	p2pListenAddr.Store(readListenAddrOrDefault(p2pListenAddrEnvVar, "127.0.0.1"))

	services := readAdditionalServices()
	validateAddionalServices(services)
	additionalServices.Store(services)
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
		case "blockchain_full", "blockchain_simple", "wallet", "miner", "app":
			services = append(services, svc)
		default:
			logger.Errorf("unknown service in %s: %s", additionalServicesEnvVar, svc)
		}
	}

	return services
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
	_, hasBlockchainSimple := seen["blockchain_simple"]

	needsBlockchain := hasWallet || hasMiner
	if needsBlockchain && !hasBlockchainFull && !hasBlockchainSimple {
		logger.Errorf("wallet or miner service requires blockchain_full or blockchain_simple to be enabled")
	}

	if hasBlockchainFull && hasBlockchainSimple {
		logger.Errorf("blockchain_full and blockchain_simple services are mutually exclusive")
	}
}

func getAdditionalServices() []string {
	return additionalServices.Load().([]string)
}

func AppPort() uint16 {
	return uint16(appPort.Load())
}

func P2PPort() uint16 {
	return uint16(p2pPort.Load())
}

func AppListenAddr() string {
	return appListenAddr.Load().(string)
}

func P2PListenAddr() string {
	return p2pListenAddr.Load().(string)
}

func P2PAdvertiseIP(bindAddr netip.Addr) netip.Addr {
	raw := strings.TrimSpace(os.Getenv(p2pAdvertiseIPEnvVar))
	if raw != "" {
		ip, err := netip.ParseAddr(raw)
		if err == nil {
			return ip
		}
		logger.Warnf("invalid %s value: %s", p2pAdvertiseIPEnvVar, raw)
	}

	if bindAddr.IsValid() && !bindAddr.IsUnspecified() {
		return bindAddr
	}

	// When binding to 0.0.0.0 in containers, pick a non-loopback IPv4 to advertise.
	if ip := firstNonLoopbackIPv4(); ip.IsValid() {
		return ip
	}

	return bindAddr
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
	return readUint16EnvOrDefault(p2pPortEnvVar, defaultP2PPort)
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

func readListenAddrOrDefault(key string, fallback string) string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	// Validate it's a plain IP address for now.
	if _, err := netip.ParseAddr(raw); err != nil {
		logger.Warnf("invalid %s value: %s, falling back to %s", key, raw, fallback)
		return fallback
	}
	return raw
}

func firstNonLoopbackIPv4() netip.Addr {
	ifaces, err := net.Interfaces()
	if err != nil {
		return netip.Addr{}
	}

	var firstIPv4 netip.Addr
	for _, iface := range ifaces {
		if (iface.Flags&net.FlagUp) == 0 || (iface.Flags&net.FlagLoopback) != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			ip := netip.Addr{}
			switch v := a.(type) {
			case *net.IPNet:
				ip, _ = netip.AddrFromSlice(v.IP)
			case *net.IPAddr:
				ip, _ = netip.AddrFromSlice(v.IP)
			}
			if !ip.IsValid() || !ip.Is4() || ip.IsLoopback() {
				continue
			}
			// Prefer private ranges.
			if ip.IsPrivate() {
				return ip
			}
			if !firstIPv4.IsValid() {
				firstIPv4 = ip
			}
		}
	}

	return firstIPv4
}

// SetAppPort sets the application port to the given value.
// Is needed for example for dynamic port assignment.
func SetAppPort(port uint16) {
	appPort.Store(uint32(port))
}

func SetP2PPort(port uint16) {
	p2pPort.Store(uint32(port))
}
