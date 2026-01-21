package adapter

import (
	"fmt"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/inv"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"
)

func ToGrpcGetDataMsg(inventory []*inv.InvVector) (*pb.GetDataMsg, error) {
	if inventory == nil {
		return nil, fmt.Errorf("inventory must not be nil")
	}
	pbInv := toGrpcInvVector(inventory)
	return &pb.GetDataMsg{Inventory: pbInv}, nil
}

func ToGrpcGetInvMsg(inventory []*inv.InvVector) (*pb.InvMsg, error) {
	if inventory == nil {
		return nil, fmt.Errorf("inventory must not be nil")
	}
	pbInv := toGrpcInvVector(inventory)
	return &pb.InvMsg{Inventory: pbInv}, nil
}

func ToGrpcBlockLocator(locator block.BlockLocator) (*pb.BlockLocator, error) {
	if locator.BlockLocatorHashes == nil {
		return nil, fmt.Errorf("block locator hashes must not be nil")
	}

	hashes := make([][]byte, len(locator.BlockLocatorHashes))
	for i, hash := range locator.BlockLocatorHashes {
		hashes[i] = hash[:]
	}

	return &pb.BlockLocator{
		BlockLocatorHashes: hashes,
		HashStop:           locator.StopHash[:],
	}, nil
}

func toGrpcInvVector(inventory []*inv.InvVector) []*pb.InvVector {
	pbInv := make([]*pb.InvVector, len(inventory))
	for i, v := range inventory {
		pbInv[i] = &pb.InvVector{
			Type: pb.InvType(v.InvType),
			Hash: v.Hash[:],
		}
	}
	return pbInv
}

func ToGrpcHeadersMsg(headers []*block.BlockHeader) (*pb.BlockHeaders, error) {
	if headers == nil {
		return nil, fmt.Errorf("headers must not be nil")
	}

	pbHeaders := make([]*pb.BlockHeader, len(headers))
	for i, h := range headers {
		if h == nil {
			return nil, fmt.Errorf("headers[%d] must not be nil", i)
		}
		pbHeaders[i] = &pb.BlockHeader{
			PrevBlockHash:    h.PreviousBlockHash[:],
			MerkleRoot:       h.MerkleRoot[:],
			Timestamp:        h.Timestamp,
			DifficultyTarget: uint32(h.DifficultyTarget),
			Nonce:            h.Nonce,
		}
	}

	return &pb.BlockHeaders{Headers: pbHeaders}, nil
}

func ToGrpcBlockMsg(b *block.Block) (*pb.BlockMsg, error) {
	if b == nil {
		return nil, fmt.Errorf("block must not be nil")
	}

	pbBlock, err := toGrpcBlock(b)
	if err != nil {
		return nil, err
	}

	return &pb.BlockMsg{Block: pbBlock}, nil
}

func toGrpcBlock(b *block.Block) (*pb.Block, error) {
	pbHeader := toGrpcBlockHeader(&b.Header)

	pbTransactions := make([]*pb.Transaction, len(b.Transactions))
	for i, tx := range b.Transactions {
		pbTx, err := toGrpcTransaction(&tx)
		if err != nil {
			return nil, fmt.Errorf("error converting transaction[%d]: %w", i, err)
		}
		pbTransactions[i] = pbTx
	}

	return &pb.Block{
		Header:       pbHeader,
		Transactions: pbTransactions,
	}, nil
}

func toGrpcBlockHeader(h *block.BlockHeader) *pb.BlockHeader {
	return &pb.BlockHeader{
		PrevBlockHash:    h.PreviousBlockHash[:],
		MerkleRoot:       h.MerkleRoot[:],
		Timestamp:        h.Timestamp,
		DifficultyTarget: uint32(h.DifficultyTarget),
		Nonce:            h.Nonce,
	}
}

func ToGrpcTxMsg(tx *transaction.Transaction) (*pb.TxMsg, error) {
	if tx == nil {
		return nil, fmt.Errorf("transaction must not be nil")
	}

	pbTx, err := toGrpcTransaction(tx)
	if err != nil {
		return nil, err
	}

	return &pb.TxMsg{Transaction: pbTx}, nil
}

func toGrpcTransaction(tx *transaction.Transaction) (*pb.Transaction, error) {
	pbInputs := make([]*pb.TxInput, len(tx.Inputs))
	for i, input := range tx.Inputs {
		pbInputs[i] = &pb.TxInput{
			PrevTxHash:  input.PrevTxID[:],
			OutputIndex: input.OutputIndex,
			Signature:   input.Signature,
			PublicKey:   input.PubKey[:],
		}
	}

	pbOutputs := make([]*pb.TxOutput, len(tx.Outputs))
	for i, output := range tx.Outputs {
		pbOutputs[i] = &pb.TxOutput{
			Value:         output.Value,
			PublicKeyHash: output.PubKeyHash[:],
		}
	}

	return &pb.Transaction{
		Inputs:  pbInputs,
		Outputs: pbOutputs,
	}, nil
}
