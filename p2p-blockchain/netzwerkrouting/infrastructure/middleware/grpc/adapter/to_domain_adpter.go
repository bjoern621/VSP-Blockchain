package adapter

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"
)

func toInvVector(pb *pb.InvVector) (block.InvVector, error) {
	return block.InvVector{}, nil
}

func ToInvVectorsFromInvMsg(pb *pb.InvMsg) ([]*block.InvVector, error) {
	return nil, nil
}

func ToInvVectorsFromGetDataMsg(pb *pb.GetDataMsg) ([]*block.InvVector, error) {
	return nil, nil
}

func ToBlockFromBlockMsg(pb *pb.BlockMsg) (block.Block, error) {
	return block.Block{}, nil
}

func ToMerkleBlockFromMerkleBlockMsg(pb *pb.MerkleBlock) (block.MerkleBlock, error) {
	return block.MerkleBlock{}, nil
}

func ToTxFromTxMsg(pb *pb.TxMsg) (transaction.Transaction, error) {
	return transaction.Transaction{}, nil
}

func ToBlockLocator(pb *pb.BlockLocator) (block.BlockLocator, error) {
	return block.BlockLocator{}, nil
}

func toHeader(pb *pb.BlockHeader) (block.BlockHeader, error) {
	return block.BlockHeader{}, nil
}

func ToHeadersFromHeadersMsg(pb *pb.BlockHeaders) ([]*block.BlockHeader, error) {
	return nil, nil
}

func ToSetFilterRequestFromSetFilterRequest(pb *pb.SetFilterRequest) (block.SetFilterRequest, error) {
	return block.SetFilterRequest{}, nil
}
