package grpc

import (
	"context"
	"fmt"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/block"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/constants"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/transaction"
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"

	"google.golang.org/protobuf/types/known/emptypb"
)

// TODO: Die ganzen Methoden in einen Adapter auslagern?
func bytesToHash(bytes []byte) (block.Hash, error) {
	var hash block.Hash
	if len(bytes) != constants.HashSize {
		return hash, fmt.Errorf("invalid hash length: %d", len(bytes))
	}

	copy(hash[:], bytes)
	return hash, nil
}

func byteListToHashes(byteList [][]byte) ([]block.Hash, error) {
	var hashes = make([]block.Hash, 0, len(byteList))
	for _, bytes := range byteList {
		hash, err := bytesToHash(bytes)
		if err != nil {
			return nil, fmt.Errorf("invalid hash length")
		}
		hashes = append(hashes, hash)
	}

	return hashes, nil
}

func bytesToPubKeyHash(bytes []byte) (transaction.PubKeyHash, error) {
	var hash transaction.PubKeyHash
	if len(bytes) != constants.PublicKeyHashSize {
		return hash, fmt.Errorf("invalid hash length: %d", len(bytes))
	}

	copy(hash[:], bytes)
	return hash, nil
}

func protoToInvVector(vector *pb.InvVector) (block.InvVector, error) {
	if vector == nil {
		return block.InvVector{}, fmt.Errorf("invalid proto inv vector")
	}

	hash, err := bytesToHash(vector.Hash)
	if err != nil {
		return block.InvVector{}, err
	}

	return block.InvVector{
		InvType: block.InvType(vector.Type),
		Hash:    hash,
	}, nil
}

func protoInvVectors(inventoryVector []*pb.InvVector) ([]block.InvVector, error) {
	if inventoryVector == nil {
		return nil, fmt.Errorf("invalid proto inv vector list")
	}

	invVectors := make([]block.InvVector, 0, len(inventoryVector))
	for _, pbInvVector := range inventoryVector {
		invVector, err := protoToInvVector(pbInvVector)
		if err != nil {
			return nil, err
		}
		invVectors = append(invVectors, invVector)
	}

	return invVectors, nil
}

func protoToTxInput(input *pb.TxInput) (transaction.Input, error) {
	if input == nil {
		return transaction.Input{}, fmt.Errorf("invalid proto tx input")
	}

	hash, err := bytesToHash(input.PrevTxHash)
	if err != nil {
		return transaction.Input{}, err
	}

	return transaction.Input{
		PrevTxID:    transaction.TransactionID(hash),
		OutputIndex: input.OutputIndex,
		Signature:   input.SignatureScript,
		Sequence:    input.Sequence,
	}, nil
}

func protoToTxInputs(pbInputs []*pb.TxInput) ([]transaction.Input, error) {
	var inputs []transaction.Input
	for _, input := range pbInputs {
		in, err := protoToTxInput(input)
		if err != nil {
			return nil, err
		}

		inputs = append(inputs, in)
	}

	return inputs, nil
}

func protoToTxOutput(output *pb.TxOutput) (transaction.Output, error) {
	if output == nil {
		return transaction.Output{}, fmt.Errorf("invalid proto tx output")
	}

	pubKeyHash, err := bytesToPubKeyHash(output.PublicKeyScript)
	if err != nil {
		return transaction.Output{}, err
	}

	return transaction.Output{
		Value:      output.Value,
		PubKeyHash: pubKeyHash,
	}, nil
}

func protoToTxOutputs(pbOutputs []*pb.TxOutput) ([]transaction.Output, error) {
	var outputs []transaction.Output
	for _, output := range pbOutputs {
		out, err := protoToTxOutput(output)
		if err != nil {
			return nil, err
		}

		outputs = append(outputs, out)
	}

	return outputs, nil
}

func protoToTransaction(tx *pb.Transaction) (transaction.Transaction, error) {
	if tx == nil {
		return transaction.Transaction{}, fmt.Errorf("invalid proto transaction")
	}

	inputs, err := protoToTxInputs(tx.Inputs)
	if err != nil {
		return transaction.Transaction{}, err
	}

	outputs, err := protoToTxOutputs(tx.Outputs)
	if err != nil {
		return transaction.Transaction{}, err
	}

	return transaction.Transaction{
		Inputs:   inputs,
		Outputs:  outputs,
		LockTime: tx.LockTime,
	}, nil
}

func protoToTransactions(pbTransactions []*pb.Transaction) ([]transaction.Transaction, error) {
	transactions := make([]transaction.Transaction, 0, len(pbTransactions))

	for _, tx := range pbTransactions {
		transactionConverted, err := protoToTransaction(tx)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, transactionConverted)
	}

	return transactions, nil
}

func protoToBlockHeader(header *pb.BlockHeader) (block.BlockHeader, error) {
	if header == nil {
		return block.BlockHeader{}, fmt.Errorf("invalid proto block header")
	}

	prevBlockHash, err := bytesToHash(header.PrevBlockHash)
	if err != nil {
		return block.BlockHeader{}, err
	}
	merkleRoot, err := bytesToHash(header.MerkleRoot)
	if err != nil {
		return block.BlockHeader{}, err
	}

	return block.BlockHeader{
		Hash:             prevBlockHash,
		MerkleRoot:       merkleRoot,
		Timestamp:        header.Timestamp,
		DifficultyTarget: header.DifficultyTarget,
		Nonce:            header.Nonce,
	}, nil
}

func protoToBlockHeaders(pbHeaders []*pb.BlockHeader) ([]block.BlockHeader, error) {
	var headers = make([]block.BlockHeader, 0, len(pbHeaders))
	for _, header := range pbHeaders {
		blockHeader, err := protoToBlockHeader(header)
		if err != nil {
			return nil, err
		}

		headers = append(headers, blockHeader)
	}

	return headers, nil
}

func protoToBlock(pbBlock *pb.Block) (block.Block, error) {
	if pbBlock == nil {
		return block.Block{}, fmt.Errorf("invalid proto block")
	}

	header, err := protoToBlockHeader(pbBlock.Header)
	if err != nil {
		return block.Block{}, err
	}

	transactions, err := protoToTransactions(pbBlock.Transactions)
	if err != nil {
		return block.Block{}, err
	}

	return block.Block{
		Header:       header,
		Transactions: transactions,
	}, nil
}

func protoToMerkleProof(pbProof *pb.MerkleProof) (block.MerkleProof, error) {
	if pbProof == nil {
		return block.MerkleProof{}, fmt.Errorf("invalid proto merkle proof")
	}

	transactionConverted, err := protoToTransaction(pbProof.Transaction)
	if err != nil {
		return block.MerkleProof{}, err
	}

	siblings, err := byteListToHashes(pbProof.Siblings)
	if err != nil {
		return block.MerkleProof{}, err
	}

	return block.MerkleProof{
		Transaction: transactionConverted,
		Siblings:    siblings,
		Index:       pbProof.Index,
	}, nil
}

func protoToMerkleProofs(pbProofs []*pb.MerkleProof) ([]block.MerkleProof, error) {
	proofs := make([]block.MerkleProof, 0, len(pbProofs))
	for _, pbProof := range pbProofs {
		proof, err := protoToMerkleProof(pbProof)
		if err != nil {
			return nil, err
		}

		proofs = append(proofs, proof)
	}

	return proofs, nil
}

func protoToMerkleBlock(merkleBlock *pb.MerkleBlock) (block.MerkleBlock, error) {
	if merkleBlock == nil {
		return block.MerkleBlock{}, fmt.Errorf("invalid proto merkle block")
	}

	header, err := protoToBlockHeader(merkleBlock.Header)
	if err != nil {
		return block.MerkleBlock{}, err
	}
	proofs, err := protoToMerkleProofs(merkleBlock.Proofs)
	if err != nil {
		return block.MerkleBlock{}, nil
	}

	return block.MerkleBlock{
		BlockHeader: header,
		Proofs:      proofs,
	}, nil
}

func protoToBlockLocator(locator *pb.BlockLocator) (block.BlockLocator, error) {
	if locator == nil {
		return block.BlockLocator{}, fmt.Errorf("invalid proto block locator")
	}

	blockLocatorHashes, err := byteListToHashes(locator.BlockLocatorHashes)
	if err != nil {
		return block.BlockLocator{}, err
	}
	hash, err := bytesToHash(locator.HashStop)
	if err != nil {
		return block.BlockLocator{}, err
	}

	return block.BlockLocator{
		BlockLocatorHashes: blockLocatorHashes,
		StopHash:           hash,
	}, nil
}

func (s *Server) Inv(ctx context.Context, msg *pb.InvMsg) (*emptypb.Empty, error) {
	inboundAddr := GetPeerAddr(ctx)
	id, suc := s.networkInfoRegistry.GetPeerIDByAddr(inboundAddr)
	if !suc {
		return nil, fmt.Errorf("peer not found for inbound address %s", inboundAddr)
	}

	invVectors, err := protoInvVectors(msg.Inventory)
	if err != nil {
		return nil, err
	}

	go s.NotifyInv(
		block.InvMsg{Inventory: invVectors},
		id,
	)

	return &emptypb.Empty{}, nil
}

func (s *Server) GetData(ctx context.Context, msg *pb.GetDataMsg) (*emptypb.Empty, error) {
	inboundAddr := GetPeerAddr(ctx)
	id, suc := s.networkInfoRegistry.GetPeerIDByAddr(inboundAddr)
	if !suc {
		return nil, fmt.Errorf("peer not found for inbound address %s", inboundAddr)
	}

	invVectors, err := protoInvVectors(msg.Inventory)
	if err != nil {
		return nil, err
	}

	go s.NotifyGetData(
		block.GetDataMsg{Inventory: invVectors},
		id,
	)

	return &emptypb.Empty{}, nil
}

func (s *Server) Block(ctx context.Context, msg *pb.BlockMsg) (*emptypb.Empty, error) {
	inboundAddr := GetPeerAddr(ctx)
	id, suc := s.networkInfoRegistry.GetPeerIDByAddr(inboundAddr)
	if !suc {
		return nil, fmt.Errorf("peer not found for inbound address %s", inboundAddr)
	}

	blockConverted, err := protoToBlock(msg.Block)
	if err != nil {
		return nil, err
	}

	go s.NotifyBlock(
		block.BlockMsg{Block: blockConverted},
		id,
	)

	return &emptypb.Empty{}, nil
}

func (s *Server) MerkleBlock(ctx context.Context, msg *pb.MerkleBlockMsg) (*emptypb.Empty, error) {
	inboundAddr := GetPeerAddr(ctx)
	id, suc := s.networkInfoRegistry.GetPeerIDByAddr(inboundAddr)
	if !suc {
		return nil, fmt.Errorf("peer not found for inbound address %s", inboundAddr)
	}

	merkleBlock, err := protoToMerkleBlock(msg.MerkleBlock)
	if err != nil {
		return nil, err
	}

	go s.NotifyMerkleBlock(
		block.MerkleBlockMsg{MerkleBlock: merkleBlock},
		id,
	)

	return &emptypb.Empty{}, nil
}

func (s *Server) Tx(ctx context.Context, msg *pb.TxMsg) (*emptypb.Empty, error) {
	inboundAddr := GetPeerAddr(ctx)
	id, suc := s.networkInfoRegistry.GetPeerIDByAddr(inboundAddr)
	if !suc {
		return nil, fmt.Errorf("peer not found for inbound address %s", inboundAddr)
	}

	transactionConverted, err := protoToTransaction(msg.Transaction)
	if err != nil {
		return nil, err
	}

	go s.NotifyTx(
		block.TxMsg{Transaction: transactionConverted},
		id,
	)

	return &emptypb.Empty{}, nil
}

func (s *Server) GetHeaders(ctx context.Context, locator *pb.BlockLocator) (*emptypb.Empty, error) {
	inboundAddr := GetPeerAddr(ctx)
	id, suc := s.networkInfoRegistry.GetPeerIDByAddr(inboundAddr)
	if !suc {
		return nil, fmt.Errorf("peer not found for inbound address %s", inboundAddr)
	}

	blockLocator, err := protoToBlockLocator(locator)
	if err != nil {
		return nil, err
	}

	go s.NotifyGetHeaders(blockLocator, id)

	return &emptypb.Empty{}, nil
}

func (s *Server) Headers(ctx context.Context, pbHeaders *pb.BlockHeaders) (*emptypb.Empty, error) {
	inboundAddr := GetPeerAddr(ctx)
	id, suc := s.networkInfoRegistry.GetPeerIDByAddr(inboundAddr)
	if !suc {
		return nil, fmt.Errorf("peer not found for inbound address %s", inboundAddr)
	}

	headers, err := protoToBlockHeaders(pbHeaders.Headers)
	if err != nil {
		return nil, err
	}

	go s.NotifyHeaders(headers, id)

	return &emptypb.Empty{}, nil
}

func (s *Server) SetFilter(ctx context.Context, request *pb.SetFilterRequest) (*emptypb.Empty, error) {
	inboundAddr := GetPeerAddr(ctx)
	id, suc := s.networkInfoRegistry.GetPeerIDByAddr(inboundAddr)
	if !suc {
		return nil, fmt.Errorf("peer not found for inbound address %s", inboundAddr)
	}

	filterRequest, err := byteListToHashes(request.PublicKeyHashes)
	if err != nil {
		return nil, err
	}

	go s.NotifySetFilterRequest(
		block.SetFilterRequest{PublicKeyHashes: filterRequest},
		id,
	)

	return &emptypb.Empty{}, nil
}

func (s *Server) Mempool(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	inboundAddr := GetPeerAddr(ctx)
	id, suc := s.networkInfoRegistry.GetPeerIDByAddr(inboundAddr)
	if !suc {
		return nil, fmt.Errorf("peer not found for inbound address %s", inboundAddr)
	}

	go s.NotifyMempool(id)

	return &emptypb.Empty{}, nil
}

func (s *Server) NotifyInv(invMsg block.InvMsg, peerID peer.PeerID) {
	for observer := range s.observers {
		observer.Inv(invMsg, peerID)
	}
}

func (s *Server) NotifyGetData(getDataMsg block.GetDataMsg, peerID peer.PeerID) {
	for observer := range s.observers {
		observer.GetData(getDataMsg, peerID)
	}
}

func (s *Server) NotifyBlock(blockMsg block.BlockMsg, peerID peer.PeerID) {
	for observer := range s.observers {
		observer.Block(blockMsg, peerID)
	}
}

func (s *Server) NotifyMerkleBlock(merkleBlockMsg block.MerkleBlockMsg, peerID peer.PeerID) {
	for observer := range s.observers {
		observer.MerkleBlock(merkleBlockMsg, peerID)
	}
}

func (s *Server) NotifyTx(txMsg block.TxMsg, peerID peer.PeerID) {
	for observer := range s.observers {
		observer.Tx(txMsg, peerID)
	}
}

func (s *Server) NotifyGetHeaders(locator block.BlockLocator, peerID peer.PeerID) {
	for observer := range s.observers {
		observer.GetHeaders(locator, peerID)
	}
}

func (s *Server) NotifyHeaders(headers []block.BlockHeader, peerID peer.PeerID) {
	for observer := range s.observers {
		observer.Headers(headers, peerID)
	}
}

func (s *Server) NotifySetFilterRequest(setFilterRequest block.SetFilterRequest, peerID peer.PeerID) {
	for observer := range s.observers {
		observer.SetFilter(setFilterRequest, peerID)
	}
}

func (s *Server) NotifyMempool(peerID peer.PeerID) {
	for observer := range s.observers {
		observer.Mempool(peerID)
	}
}
