package mapping

import (
	"net/netip"

	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer/discovery"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/infrastructure/middleware/grpc/networkinfo"
)

// PeerAddressFromProto converts protobuf PeerAddress (IP + port + timestamp) to domain PeerAddress (PeerId + timestamp).
// The infrastructure layer looks up or registers the PeerId from the IP address.
func PeerAddressFromProto(pbPeerAddr *pb.PeerAddress, registry *networkinfo.NetworkInfoRegistry) (discovery.PeerAddress, error) {
	if pbPeerAddr == nil || pbPeerAddr.ListeningEndpoint == nil {
		return discovery.PeerAddress{}, nil
	}

	ip, ok := netip.AddrFromSlice(pbPeerAddr.ListeningEndpoint.IpAddress)
	if !ok {
		return discovery.PeerAddress{}, nil
	}

	addrPort := netip.AddrPortFrom(ip, uint16(pbPeerAddr.ListeningEndpoint.ListeningPort))

	// Look up or register the peer by their listening endpoint
	// The registry handles the mapping between IP addresses and PeerIds
	peerID := registry.GetOrRegisterPeer(netip.AddrPort{}, addrPort)

	return discovery.PeerAddress{
		PeerId:              peerID,
		LastActiveTimestamp: pbPeerAddr.LastActiveTimestamp,
	}, nil
}

// PeerAddressToProto converts domain PeerAddress (PeerId + timestamp) to protobuf PeerAddress (IP + port + timestamp).
// The infrastructure layer looks up the listening endpoint for the given PeerId.
func PeerAddressToProto(peerAddr discovery.PeerAddress, registry *networkinfo.NetworkInfoRegistry) *pb.PeerAddress {
	listeningEndpoint, ok := registry.GetListeningEndpoint(peerAddr.PeerId)
	if !ok {
		// Peer has no listening endpoint, return an empty PeerAddress
		return &pb.PeerAddress{
			ListeningEndpoint:   nil,
			LastActiveTimestamp: peerAddr.LastActiveTimestamp,
		}
	}

	return &pb.PeerAddress{
		ListeningEndpoint: &pb.Endpoint{
			IpAddress:     listeningEndpoint.Addr().AsSlice(),
			ListeningPort: uint32(listeningEndpoint.Port()),
		},
		LastActiveTimestamp: peerAddr.LastActiveTimestamp,
	}
}
