package common

import (
	"fmt"
	"net/netip"
	"sync/atomic"
)

const (
	defaultP2PPort = 50051
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
