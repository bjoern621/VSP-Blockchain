package grpc

import (
	"context"

	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer/discovery"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/infrastructure/middleware/grpc/mapping"

	"bjoernblessin.de/go-utils/util/logger"
	"google.golang.org/protobuf/types/known/emptypb"
)

// GetAddr handles incoming GetAddr requests from peers.
// This method is called when another peer requests our known peer addresses.
// It delegates to the discovery service to retrieve and send all known peer addresses.
func (s *Server) GetAddr(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	peerId := s.GetPeerId(ctx)

	s.discoveryService.HandleGetAddr(peerId)

	return &emptypb.Empty{}, nil
}

// Addr handles incoming Addr messages from peers.
// This method is called when another peer shares known peer addresses with us.
// It delegates to the discovery service to process the received addresses.
func (s *Server) Addr(ctx context.Context, req *pb.AddrList) (*emptypb.Empty, error) {
	peerId := s.GetPeerId(ctx)

	// Convert protobuf addresses (IP + port + timestamp) to domain model (PeerId + timestamp)
	// The infrastructure layer looks up or registers PeerIds from IP addresses
	peerAddresses := make([]discovery.PeerAddress, 0, len(req.Peers))
	for _, peerAddr := range req.Peers {
		pa, err := mapping.PeerAddressFromProto(peerAddr, s.networkInfoRegistry)
		if err != nil {
			logger.Warnf("[peerdiscovery_grpc] Failed to parse peer address for peer %s: %v", peerId, err)
			continue
		}
		peerAddresses = append(peerAddresses, pa)
	}

	s.discoveryService.HandleAddr(peerId, peerAddresses)

	return &emptypb.Empty{}, nil
}

// SendGetAddr sends a GetAddr request to the specified peer.
// This method is used to request known peer addresses from a specific peer.
func (c *Client) SendGetAddr(peerID common.PeerId) {
	SendHelper(c, peerID, "GetAddr", pb.NewPeerDiscoveryClient, func(client pb.PeerDiscoveryClient) error {
		_, err := client.GetAddr(context.Background(), &emptypb.Empty{})
		return err
	})
}

// SendAddr sends an Addr message containing known peer addresses to the specified peer.
// This method is used to share known peer addresses with a specific peer.
func (c *Client) SendAddr(peerID common.PeerId, peerAddresses []discovery.PeerAddress) {
	// Convert domain model (PeerId + timestamp) to protobuf addresses (IP + port + timestamp)
	// The infrastructure layer looks up listening endpoints for PeerIds
	pbPeers := make([]*pb.PeerAddress, 0, len(peerAddresses))
	for _, peerAddr := range peerAddresses {
		pbPeer := mapping.PeerAddressToProto(peerAddr, c.networkInfoRegistry)
		// Skip peers without listening endpoints
		if pbPeer.ListeningEndpoint == nil {
			continue
		}
		pbPeers = append(pbPeers, pbPeer)
	}

	addrList := &pb.AddrList{
		Peers: pbPeers,
	}

	SendHelper(c, peerID, "Addr", pb.NewPeerDiscoveryClient, func(client pb.PeerDiscoveryClient) error {
		_, err := client.Addr(context.Background(), addrList)
		return err
	})
}
