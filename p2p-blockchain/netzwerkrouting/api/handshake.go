package api

import "net/netip"

type HandshakeAPI interface {
	InitiateHandshake(addrPort netip.AddrPort)
}
