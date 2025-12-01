package common

import (
	"fmt"
	"net/netip"
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
	P2PListeningIpAddr netip.Addr
)

func SetP2PListeningIpAddr(ip netip.Addr) {
	P2PListeningIpAddr = ip
}
