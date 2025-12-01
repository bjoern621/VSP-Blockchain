package common

import "net/netip"

const (
	defaultP2PPort = 50051
	defaultAppPort = 50050
)

var (
	P2PListeningIpAddr netip.Addr
)

func SetP2PListeningIpAddr(ip netip.Addr) {
	P2PListeningIpAddr = ip
}
