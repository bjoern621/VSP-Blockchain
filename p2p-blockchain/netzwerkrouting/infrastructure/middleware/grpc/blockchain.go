package grpc

import (
	"context"
	"fmt"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain"
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"

	"bjoernblessin.de/go-utils/util/assert"
	"google.golang.org/protobuf/types/known/emptypb"
)

// TODO: Die ganzen Methoden in einen Adapter auslagern?
func bytesToHash(bytes []byte) (*blockchain.Hash, error) {
	var hash blockchain.Hash
	if len(bytes) != 32 {
		return &hash, fmt.Errorf("invalid hash length: %d", len(bytes))
	}

	copy(hash[:], bytes)
	return &hash, nil
}

func byteListToHashes(byteList [][]byte) []*blockchain.Hash {
	var hashes = make([]*blockchain.Hash, len(byteList))
	for _, bytes := range byteList {
		hash, err := bytesToHash(bytes)
		assert.IsNil(err, "invalid hash length")
		hashes = append(hashes, hash)
	}

	return hashes
}

func protoToInvVector(vector *pb.InvVector) *blockchain.InvVector {
	if vector == nil {
		return nil
	}

	hash, err := bytesToHash(vector.Hash)
	assert.IsNil(err, "failed to convert hash from proto")

	return &blockchain.InvVector{
		InvType: blockchain.InvType(vector.Type),
		Hash:    hash,
	}
}

func protoInvVectors(inventoryVector []*pb.InvVector) []*blockchain.InvVector {
	if inventoryVector == nil {
		return nil
	}

	invVectors := make([]*blockchain.InvVector, len(inventoryVector))
	for _, pbInvVector := range inventoryVector {
		invVectors = append(invVectors, protoToInvVector(pbInvVector))
	}

	return invVectors
}

func protoToTxInput(input *pb.TxInput) *blockchain.TxInput {
	if input == nil {
		return nil
	}

	hash, err := bytesToHash(input.PrevTxHash)
	assert.IsNil(err, "failed to convert previous tx hash from proto")

	return &blockchain.TxInput{
		PreviousTxHash:  hash,
		PreviousIndex:   input.OutputIndex,
		SignatureScript: input.SignatureScript,
		Sequence:        input.Sequence,
	}
}

func protoToTxInputs(pbInputs []*pb.TxInput) []*blockchain.TxInput {
	var inputs []*blockchain.TxInput
	for _, input := range pbInputs {
		inputs = append(inputs, protoToTxInput(input))
	}

	return inputs
}

func protoToTxOutput(output *pb.TxOutput) *blockchain.TxOutput {
	if output == nil {
		return nil
	}

	return &blockchain.TxOutput{
		Value:           output.Value,
		PublicKeyScript: output.PublicKeyScript,
	}
}

func protoToTxOutputs(pbOutputs []*pb.TxOutput) []*blockchain.TxOutput {
	var outputs []*blockchain.TxOutput
	for _, output := range pbOutputs {
		outputs = append(outputs, protoToTxOutput(output))
	}

	return outputs
}

func protoToTransaction(tx *pb.Transaction) *blockchain.Transaction {
	if tx == nil {
		return nil
	}

	return &blockchain.Transaction{
		Inputs:   protoToTxInputs(tx.Inputs),
		Outputs:  protoToTxOutputs(tx.Outputs),
		LockTime: tx.LockTime,
	}
}

func protoToTransactions(pbTransactions []*pb.Transaction) []*blockchain.Transaction {
	transactions := make([]*blockchain.Transaction, len(pbTransactions))
	for _, tx := range pbTransactions {
		transactions = append(transactions, protoToTransaction(tx))
	}

	return transactions
}

func protoToBlockHeader(header *pb.BlockHeader) *blockchain.BlockHeader {
	if header == nil {
		return nil
	}

	prevBlockHash, err := bytesToHash(header.PrevBlockHash)
	assert.IsNil(err, "failed to convert previous block hash from proto")
	merkleRoot, err := bytesToHash(header.MerkleRoot)
	assert.IsNil(err, "failed to convert merkle root hash from proto")

	return &blockchain.BlockHeader{
		Hash:             prevBlockHash,
		MerkleRoot:       merkleRoot,
		Timestamp:        header.Timestamp,
		DifficultyTarget: header.DifficultyTarget,
		Nonce:            header.Nonce,
	}
}

func protoToBlockHeaders(pbHeaders []*pb.BlockHeader) []*blockchain.BlockHeader {
	var headers = make([]*blockchain.BlockHeader, len(pbHeaders))
	for _, header := range pbHeaders {
		headers = append(headers, protoToBlockHeader(header))
	}

	return headers
}

func protoToBlock(block *pb.Block) *blockchain.Block {
	if block == nil {
		return nil
	}

	return &blockchain.Block{
		Header:       protoToBlockHeader(block.Header),
		Transactions: protoToTransactions(block.Transactions),
	}
}

func protoToMerkleProof(pbProof *pb.MerkleProof) *blockchain.MerkleProof {
	if pbProof == nil {
		return nil
	}

	return &blockchain.MerkleProof{
		Transaction: protoToTransaction(pbProof.Transaction),
		Siblings:    byteListToHashes(pbProof.Siblings),
		Index:       pbProof.Index,
	}
}

func protoToMerkleProofs(pbProofs []*pb.MerkleProof) []*blockchain.MerkleProof {
	proofs := make([]*blockchain.MerkleProof, len(pbProofs))
	for _, proof := range pbProofs {
		proofs = append(proofs, protoToMerkleProof(proof))
	}

	return proofs
}

func protoToMerkleBlock(block *pb.MerkleBlock) *blockchain.MerkleBlock {
	if block == nil {
		return nil
	}

	return &blockchain.MerkleBlock{
		BlockHeader: protoToBlockHeader(block.Header),
		Proofs:      protoToMerkleProofs(block.Proofs),
	}
}

func protoToBlockLocator(locator *pb.BlockLocator) *blockchain.BlockLocator {
	if locator == nil {
		return nil
	}

	hash, err := bytesToHash(locator.HashStop)
	assert.IsNil(err, "error converting stop hash")

	return &blockchain.BlockLocator{
		BlockLocatorHashes: byteListToHashes(locator.BlockLocatorHashes),
		StopHash:           hash,
	}
}

func (s *Server) Inv(ctx context.Context, msg *pb.InvMsg) (*emptypb.Empty, error) {
	invVectors := protoInvVectors(msg.Inventory)
	go s.NotifyInv(&blockchain.InvMsg{
		Inventory: invVectors,
	})

	return &emptypb.Empty{}, nil
}

func (s *Server) GetData(ctx context.Context, msg *pb.GetDataMsg) (*emptypb.Empty, error) {
	invVectors := protoInvVectors(msg.Inventory)
	go s.NotifyGetData(&blockchain.GetDataMsg{
		Inventory: invVectors,
	})

	return &emptypb.Empty{}, nil
}

func (s *Server) Block(ctx context.Context, msg *pb.BlockMsg) (*emptypb.Empty, error) {
	block := protoToBlock(msg.Block)
	go s.NotifyBlock(&blockchain.BlockMsg{
		Block: block,
	})

	return &emptypb.Empty{}, nil
}

func (s *Server) MerkleBlock(ctx context.Context, msg *pb.MerkleBlockMsg) (*emptypb.Empty, error) {
	merkleBlock := protoToMerkleBlock(msg.MerkleBlock)
	go s.NotifyMerkleBlock(&blockchain.MerkleBlockMsg{
		MerkleBlock: merkleBlock,
	})

	return &emptypb.Empty{}, nil
}

func (s *Server) Tx(ctx context.Context, msg *pb.TxMsg) (*emptypb.Empty, error) {
	transaction := protoToTransaction(msg.Transaction)
	go s.NotifyTx(&blockchain.TxMsg{
		Transaction: transaction,
	})

	return &emptypb.Empty{}, nil
}

func (s *Server) GetHeaders(ctx context.Context, locator *pb.BlockLocator) (*emptypb.Empty, error) {
	blockLocator := protoToBlockLocator(locator)
	go s.NotifyGetHeaders(blockLocator)

	return &emptypb.Empty{}, nil
}

func (s *Server) Headers(ctx context.Context, pbHeaders *pb.BlockHeaders) (*emptypb.Empty, error) {
	headers := protoToBlockHeaders(pbHeaders.Headers)
	go s.NotifyHeaders(headers)

	return &emptypb.Empty{}, nil
}

func (s *Server) SetFilter(ctx context.Context, request *pb.SetFilterRequest) (*emptypb.Empty, error) {
	filterRequest := byteListToHashes(request.PublicKeyHashes)
	go s.NotifySetFilterRequest(&blockchain.SetFilterRequest{
		PublicKeyHashes: filterRequest,
	})

	return &emptypb.Empty{}, nil
}

func (s *Server) Mempool(ctx context.Context, empty *emptypb.Empty) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Server) NotifyInv(invMsg *blockchain.InvMsg) {
	for observer := range s.observers {
		observer.Inv(invMsg)
	}
}

func (s *Server) NotifyGetData(getDataMsg *blockchain.GetDataMsg) {
	for observer := range s.observers {
		observer.GetData(getDataMsg)
	}
}

func (s *Server) NotifyBlock(blockMsg *blockchain.BlockMsg) {
	for observer := range s.observers {
		observer.Block(blockMsg)
	}
}

func (s *Server) NotifyMerkleBlock(merkleBlockMsg *blockchain.MerkleBlockMsg) {
	for observer := range s.observers {
		observer.MerkleBlock(merkleBlockMsg)
	}
}

func (s *Server) NotifyTx(txMsg *blockchain.TxMsg) {
	for observer := range s.observers {
		observer.Tx(txMsg)
	}
}

func (s *Server) NotifyGetHeaders(locator *blockchain.BlockLocator) {
	for observer := range s.observers {
		observer.GetHeaders(locator)
	}
}

func (s *Server) NotifyHeaders(headers []*blockchain.BlockHeader) {
	for observer := range s.observers {
		observer.Headers(headers)
	}
}

func (s *Server) NotifySetFilterRequest(setFilterRequest *blockchain.SetFilterRequest) {
	for observer := range s.observers {
		observer.SetFilter(setFilterRequest)
	}
}
