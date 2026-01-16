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
func (s *Server) GetAddr(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	peerId := s.GetPeerId(ctx)
	logger.Debugf("Received GetAddr from peer %s", peerId)

	s.discoveryService.HandleGetAddr(peerId)

	return &emptypb.Empty{}, nil
}

// Addr handles incoming Addr messages from peers.
// This method is called when another peer shares known peer addresses with us.
// It delegates to the discovery service to process the received addresses.
func (s *Server) Addr(ctx context.Context, req *pb.AddrList) (*emptypb.Empty, error) {
	peerId := s.GetPeerId(ctx)
	logger.Debugf("Received Addr from peer %s with %d addresses", peerId, len(req.Peers))

	// Convert protobuf addresses (IP + port + timestamp) to domain model (PeerId + timestamp)
	// The infrastructure layer looks up or registers PeerIds from IP addresses
	peerAddresses := make([]discovery.PeerAddress, 0, len(req.Peers))
	for _, peerAddr := range req.Peers {
		pa, err := mapping.PeerAddressFromProto(peerAddr, s.networkInfoRegistry)
		if err != nil {
			logger.Warnf("Failed to parse peer address for peer %s: %v", peerId, err)
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
	remoteAddrPort, ok := c.networkInfoRegistry.GetListeningEndpoint(peerID)
	if !ok {
		logger.Warnf("failed to send GetAddr: no listening endpoint for peer %s", peerID)
		return
	}

	conn, ok := c.networkInfoRegistry.GetConnection(peerID)
	if !ok {
		// Create a new connection if one doesn't exist
		newConn, err := createGRPCClient(remoteAddrPort)
		if err != nil {
			logger.Warnf("failed to create gRPC client for %s: %v", remoteAddrPort.String(), err)
			return
		}
		c.networkInfoRegistry.SetConnection(peerID, newConn)
		conn = newConn
	}

	client := pb.NewPeerDiscoveryClient(conn)

	_, err := client.GetAddr(context.Background(), &emptypb.Empty{})
	if err != nil {
		logger.Warnf("failed to send GetAddr to %s: %v", peerID, err)
	}
}

// SendAddr sends an Addr message containing known peer addresses to the specified peer.
// This method is used to share known peer addresses with a specific peer.
func (c *Client) SendAddr(peerID common.PeerId, peerAddresses []discovery.PeerAddress) {
	remoteAddrPort, ok := c.networkInfoRegistry.GetListeningEndpoint(peerID)
	if !ok {
		logger.Warnf("failed to send Addr: no listening endpoint for peer %s", peerID)
		return
	}

	conn, ok := c.networkInfoRegistry.GetConnection(peerID)
	if !ok {
		// Create a new connection if one doesn't exist
		newConn, err := createGRPCClient(remoteAddrPort)
		if err != nil {
			logger.Warnf("failed to create gRPC client for %s: %v", remoteAddrPort.String(), err)
			return
		}
		c.networkInfoRegistry.SetConnection(peerID, newConn)
		conn = newConn
	}

	client := pb.NewPeerDiscoveryClient(conn)

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

	if len(pbPeers) == 0 {
		logger.Warnf("No peers with listening endpoints to send to %s", peerID)
		return
	}

	addrList := &pb.AddrList{
		Peers: pbPeers,
	}

	_, err := client.Addr(context.Background(), addrList)
	if err != nil {
		logger.Warnf("failed to send Addr to %s: %v", peerID, err)
	}
}

// HeartbeatPing handles incoming HeartbeatPing messages from peers.
// This method is called when another peer sends a ping to check liveness.
// It updates the peer's LastSeen timestamp and responds with HeartbeatPong.
func (s *Server) HeartbeatPing(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	peerId := s.GetPeerId(ctx)
	logger.Debugf("Received HeartbeatPing from peer %s", peerId)

	s.keepaliveService.HandleHeartbeatPing(peerId)

	return &emptypb.Empty{}, nil
}

// HeartbeatPong handles incoming HeartbeatPong messages from peers.
// This method is called when another peer responds to our HeartbeatPing.
// It updates the peer's LastSeen timestamp.
func (s *Server) HeartbeatPong(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	peerId := s.GetPeerId(ctx)
	logger.Debugf("Received HeartbeatPong from peer %s", peerId)

	s.keepaliveService.HandleHeartbeatPong(peerId)

	return &emptypb.Empty{}, nil
}

// SendHeartbeatPing sends a HeartbeatPing message to the specified peer and waits for a HeartbeatPong response.
// This method is used as a keep-alive mechanism to check if a peer is still responsive.
func (c *Client) SendHeartbeatPing(peerID common.PeerId) {
	conn, ok := c.networkInfoRegistry.GetConnection(peerID)
	if !ok {
		// If there's no connection, try to create one using the listening endpoint
		remoteAddrPort, ok := c.networkInfoRegistry.GetListeningEndpoint(peerID)
		if !ok {
			logger.Warnf("failed to send HeartbeatPing: no connection or listening endpoint for peer %s", peerID)
			return
		}

		newConn, err := createGRPCClient(remoteAddrPort)
		if err != nil {
			logger.Warnf("failed to create gRPC client for %s: %v", remoteAddrPort.String(), err)
			return
		}
		c.networkInfoRegistry.SetConnection(peerID, newConn)
		conn = newConn
	}

	client := pb.NewPeerDiscoveryClient(conn)

	_, err := client.HeartbeatPing(context.Background(), &emptypb.Empty{})
	if err != nil {
		logger.Warnf("failed to send HeartbeatPing to %s: %v", peerID, err)
	}
}

// SendHeartbeatPong sends a HeartbeatPong message to the specified peer.
// This method is used to respond to a HeartbeatPing from another peer.
func (c *Client) SendHeartbeatPong(peerID common.PeerId) {
	conn, ok := c.networkInfoRegistry.GetConnection(peerID)
	if !ok {
		logger.Warnf("failed to send HeartbeatPong: no connection for peer %s", peerID)
		return
	}

	client := pb.NewPeerDiscoveryClient(conn)

	_, err := client.HeartbeatPong(context.Background(), &emptypb.Empty{})
	if err != nil {
		logger.Warnf("failed to send HeartbeatPong to %s: %v", peerID, err)
	}
}
