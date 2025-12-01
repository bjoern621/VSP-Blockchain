package grpc

import "s3b/vsp-blockchain/p2p-blockchain/internal/pb"

type Client struct {
	grpcClient pb.ConnectionEstablishmentClient
}
