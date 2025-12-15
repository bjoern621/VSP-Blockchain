package common

import (
	"fmt"
	"net/netip"
	"slices"
	"sync/atomic"
)

const (
	DefaultP2PPort = 50051
	defaultAppPort = 50050
	VersionNumber  = 1
)

var (
	VersionString = fmt.Sprintf("vsgoin-%d.0", VersionNumber)
)

var (
	p2pListeningIpAddr atomic.Value // netip.Addr
)

func init() {
	p2pListeningIpAddr.Store(netip.Addr{})
}

func P2PListeningIpAddr() netip.Addr {
	return p2pListeningIpAddr.Load().(netip.Addr)
}

func SetP2PListeningIpAddr(ip netip.Addr) {
	p2pListeningIpAddr.Store(ip)
}

// EnabledTeilsystemeNames returns the names of all enabled subsystems.
// I.e. ["blockchain_full", "wallet", ...].
// Never includes "app" (which is not a subsystem).
// Always includes "netzwerkrouting".
// All available subsystems are: "blockchain_full", "blockchain_simple", "wallet", "miner", "netzwerkrouting".
// "blockchain_full" and "blockchain_simple" are mutually exclusive.
func EnabledTeilsystemeNames() []string {
	services := getAdditionalServices()
	services = slices.DeleteFunc(services, func(s string) bool { return s == "app" })
	services = append(services, "netzwerkrouting")
	return services
}

func WalletEnabled() bool {
	return slices.Contains(getAdditionalServices(), "wallet")
}

func MinerEnabled() bool {
	return slices.Contains(getAdditionalServices(), "miner")
}

func BlockchainFullEnabled() bool {
	return slices.Contains(getAdditionalServices(), "blockchain_full")
}

func BlockchainSimpleEnabled() bool {
	return slices.Contains(getAdditionalServices(), "blockchain_simple")
}

func AppEnabled() bool {
	return slices.Contains(getAdditionalServices(), "app")
}
