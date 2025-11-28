package config

import (
	"net/netip"
	"sync"
)

var (
	mu          sync.RWMutex
	p2pPort     uint16
	localIP     netip.Addr
	initialized bool
)

// Init initializes the configuration with the given values.
func Init(port uint16, ip netip.Addr) {
	mu.Lock()
	defer mu.Unlock()
	p2pPort = port
	localIP = ip
	initialized = true
}

// GetP2PPort returns the configured P2P listening port.
func GetP2PPort() uint16 {
	mu.RLock()
	defer mu.RUnlock()
	return p2pPort
}

// GetLocalIP returns the configured local IP address.
// Returns nil if not set.
func GetLocalIP() netip.Addr {
	mu.RLock()
	defer mu.RUnlock()
	return localIP
}

// GetLocalIPBytes returns the local IP as a byte slice.
// Returns 0.0.0.0 if not set.
func GetLocalIPBytes() []byte {
	mu.RLock()
	defer mu.RUnlock()

	return localIP.AsSlice()
}

// IsInitialized returns true if the configuration has been initialized.
func IsInitialized() bool {
	mu.RLock()
	defer mu.RUnlock()
	return initialized
}
