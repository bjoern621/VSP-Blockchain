package registry

import (
	"context"
	"net"
	"net/netip"

	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/infrastructure/middleware/grpc/networkinfo"
)

// dnsRegistryQuerier implements peer.RegistryQuerier and api.FullRegistryQuerier using DNS lookups.
type dnsRegistryQuerier struct {
	networkInfoRegistry *networkinfo.NetworkInfoRegistry
}

func NewDNSRegistryQuerier(networkInfoRegistry *networkinfo.NetworkInfoRegistry) peer.RegistryQuerier {
	return newDNSRegistryQuerier(networkInfoRegistry)
}

func NewDNSFullRegistryQuerier(networkInfoRegistry *networkinfo.NetworkInfoRegistry) api.FullRegistryQuerier {
	return newDNSRegistryQuerier(networkInfoRegistry)
}

func newDNSRegistryQuerier(networkInfoRegistry *networkinfo.NetworkInfoRegistry) *dnsRegistryQuerier {
	return &dnsRegistryQuerier{
		networkInfoRegistry: networkInfoRegistry,
	}
}

// QueryPeers queries the DNS seed registry and returns a list of peer IDs.
func (r *dnsRegistryQuerier) QueryPeers() ([]peer.PeerID, error) {
	entries, err := r.queryRegistry()
	if err != nil {
		return nil, err
	}

	peers := make([]peer.PeerID, 0, len(entries))
	for _, entry := range entries {
		peers = append(peers, entry.PeerID)
	}

	return peers, nil
}

// QueryFullRegistry queries the DNS seed registry and returns full entries including IP addresses.
func (r *dnsRegistryQuerier) QueryFullRegistry() ([]api.RegistryEntry, error) {
	return r.queryRegistry()
}

// queryRegistry queries the DNS seed registry for available peer addresses.
func (r *dnsRegistryQuerier) queryRegistry() ([]api.RegistryEntry, error) {
	addrs, err := net.DefaultResolver.LookupNetIP(context.Background(), "ip4", common.RegistrySeedHostnameEnv())
	if err != nil {
		return nil, err
	}

	entries := make([]api.RegistryEntry, 0, len(addrs))
	for _, addr := range addrs {
		addrPort := netip.AddrPortFrom(addr, common.DefaultP2PPort)

		peerID := r.networkInfoRegistry.GetOrRegisterPeer(netip.AddrPort{}, addrPort)
		entries = append(entries, api.RegistryEntry{
			IPAddress: addr,
			PeerID:    peerID,
		})
	}

	return entries, nil
}
