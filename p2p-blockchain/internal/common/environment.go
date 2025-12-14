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
	appPortEnvVar        = "APP_PORT"
	p2pPortEnvVar        = "P2P_PORT"
	appListenAddrEnvVar  = "APP_LISTEN_ADDR"
	p2pListenAddrEnvVar  = "P2P_LISTEN_ADDR"
	p2pAdvertiseIPEnvVar = "P2P_ADVERTISE_IP"
)

var (
	appPort       atomic.Uint32
	p2pPort       atomic.Uint32
	appListenAddr atomic.Value // string
	p2pListenAddr atomic.Value // string
)

// init reads all required environment variables at startup.
// Read values are stored in package-level variables for easy access.
func init() {
	appPort.Store(uint32(readAppPort()))
	p2pPort.Store(uint32(readP2PPort()))
	appListenAddr.Store(readListenAddrOrDefault(appListenAddrEnvVar, "127.0.0.1"))
	p2pListenAddr.Store(readListenAddrOrDefault(p2pListenAddrEnvVar, "127.0.0.1"))
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
	portStr, found := env.ReadOptionalEnv(appPortEnvVar)

	if !found {
		return defaultAppPort
	}

	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		logger.Errorf("invalid PORT value: %s, must be between 0 and 65535", portStr)
	}

	assert.Assert(port <= math.MaxUint16, "port value %d out of range", port)

	return uint16(port)
}

// readP2PPort is similar to readAppPort but for the P2P port.
func readP2PPort() uint16 {
	portStr, found := env.ReadOptionalEnv(p2pPortEnvVar)

	if !found {
		return defaultP2PPort
	}

	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		logger.Errorf("invalid P2P_PORT value: %s, must be between 0 and 65535", portStr)
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
