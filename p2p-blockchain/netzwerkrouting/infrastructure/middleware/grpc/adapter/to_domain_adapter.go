package adapter

import (
	"fmt"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/inv"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"
)

func toInvVector(pbVec *pb.InvVector) (inv.InvVector, error) {
	if pbVec == nil {
		return inv.InvVector{}, fmt.Errorf("inv vector must not be nil")
	}
	if len(pbVec.Hash) != common.HashSize {
		return inv.InvVector{}, fmt.Errorf("invalid hash length: %d", len(pbVec.Hash))
	}

	var hash common.Hash
	copy(hash[:], pbVec.Hash)
	return inv.InvVector{
		InvType: inv.InvType(pbVec.Type),
		Hash:    hash,
	}, nil
}

func ToInvVectorsFromInvMsg(pbMsg *pb.InvMsg) ([]*inv.InvVector, error) {
	if pbMsg == nil {
		return nil, fmt.Errorf("inv msg must not be nil")
	}
	if pbMsg.Inventory == nil {
		return nil, fmt.Errorf("inv msg inventory must not be nil")
	}

	return toInvVectors(pbMsg.Inventory)
}

func ToInvVectorsFromGetDataMsg(pb *pb.GetDataMsg) ([]*inv.InvVector, error) {
	if pb == nil {
		return nil, fmt.Errorf("getDataMsg must not be nil")
	}
	if pb.Inventory == nil {
		return nil, fmt.Errorf("getDataMsg inventory must not be nil")
	}

	return toInvVectors(pb.Inventory)
}

func ToBlockFromBlockMsg(pbMsg *pb.BlockMsg) (block.Block, error) {
	if pbMsg == nil || pbMsg.Block == nil {
		return block.Block{}, fmt.Errorf("block msg must not be nil")
	}
	header, err := toHeader(pbMsg.Block.Header)
	if err != nil {
		return block.Block{}, err
	}

	if pbMsg.Block.Transactions == nil {
		return block.Block{}, fmt.Errorf("block msg transactions must not be nil")
	}

	transactions, err := toTransactions(pbMsg.Block.Transactions)
	if err != nil {
		return block.Block{}, err
	}

	return block.Block{
		Header:       header,
		Transactions: transactions,
	}, nil
}

func ToMerkleBlockFromMerkleBlockMsg(pbMsg *pb.MerkleBlockMsg) (block.MerkleBlock, error) {
	if pbMsg == nil || pbMsg.MerkleBlock == nil {
		return block.MerkleBlock{}, fmt.Errorf("merkle block msg must not be nil")
	}

	header, err := toHeader(pbMsg.MerkleBlock.Header)
	if err != nil {
		return block.MerkleBlock{}, err
	}

	proofs, err := toMerkleProofs(pbMsg.MerkleBlock.Proofs)
	if err != nil {
		return block.MerkleBlock{}, err
	}

	return block.MerkleBlock{
		BlockHeader: header,
		Proofs:      proofs,
	}, nil
}

func ToTxFromTxMsg(pbMsg *pb.TxMsg) (transaction.Transaction, error) {
	if pbMsg == nil {
		return transaction.Transaction{}, fmt.Errorf("tx msg must not be nil")
	}
	tx, err := toTx(pbMsg.Transaction)
	if err != nil {
		return transaction.Transaction{}, err
	}
	return tx, nil
}

func ToBlockLocator(pb *pb.BlockLocator) (block.BlockLocator, error) {
	if pb == nil {
		return block.BlockLocator{}, fmt.Errorf("block locator must not be nil")
	}
	if pb.BlockLocatorHashes == nil {
		return block.BlockLocator{}, fmt.Errorf("block locator hashes must not be nil")
	}

	hashes := make([]common.Hash, len(pb.BlockLocatorHashes))
	for i, h := range pb.BlockLocatorHashes {
		var hash common.Hash
		copy(hash[:], h[:])
		hashes[i] = hash
	}

	if pb.HashStop == nil {
		return block.BlockLocator{}, fmt.Errorf("block locator hash stop must not be nil")
	}
	var stopHash common.Hash
	copy(stopHash[:], pb.HashStop[:])

	return block.BlockLocator{
		BlockLocatorHashes: hashes,
		StopHash:           stopHash,
	}, nil
}

func ToHeadersFromHeadersMsg(pbMsg *pb.BlockHeaders) ([]*block.BlockHeader, error) {
	if pbMsg == nil {
		return nil, fmt.Errorf("block headers must not be nil")
	}
	if pbMsg.Headers == nil {
		return nil, fmt.Errorf("block headers headers must not be nil")
	}

	headers := make([]*block.BlockHeader, len(pbMsg.Headers))
	for i, h := range pbMsg.Headers {
		if h == nil {
			return nil, fmt.Errorf("block headers headers[%d] must not be nil", i)
		}
		header, err := toHeader(h)
		if err != nil {
			return nil, err
		}
		headers[i] = &header
	}

	return headers, nil
}

func ToSetFilterRequestFromSetFilterRequest(pb *pb.SetFilterRequest) (block.SetFilterRequest, error) {
	if pb == nil {
		return block.SetFilterRequest{}, fmt.Errorf("set filter request must not be nil")
	}
	if pb.PublicKeyHashes == nil {
		return block.SetFilterRequest{}, fmt.Errorf("set filter request public key hashes must not be nil")
	}

	hashes := make([]block.PublicKeyHash, len(pb.PublicKeyHashes))
	for i, h := range pb.PublicKeyHashes {
		if h == nil {
			return block.SetFilterRequest{}, fmt.Errorf("set filter request public key hashes[%d] must not be nil", i)
		}
		if len(h) != common.HashSize {
			return block.SetFilterRequest{}, fmt.Errorf("set filter request public key hashes[%d] must be %d bytes long", i, common.HashSize)
		}

		var hash block.PublicKeyHash
		copy(hash[:], h[:])
		hashes[i] = hash
	}

	return block.SetFilterRequest{
		PublicKeyHashes: hashes,
	}, nil
}

func toHeader(pb *pb.BlockHeader) (block.BlockHeader, error) {
	if pb == nil {
		return block.BlockHeader{}, fmt.Errorf("block header must not be nil")
	}
	if len(pb.PrevBlockHash) != common.HashSize {
		return block.BlockHeader{}, fmt.Errorf("invalid prev block hash length: %d", len(pb.PrevBlockHash))
	}
	if len(pb.MerkleRoot) != common.HashSize {
		return block.BlockHeader{}, fmt.Errorf("invalid merkle root length: %d", len(pb.MerkleRoot))
	}

	var prevBlockHash common.Hash
	copy(prevBlockHash[:], pb.PrevBlockHash)

	var merkleRoot common.Hash
	copy(merkleRoot[:], pb.MerkleRoot)

	return block.BlockHeader{
		PreviousBlockHash: prevBlockHash,
		MerkleRoot:        merkleRoot,
		Timestamp:         pb.Timestamp,
		DifficultyTarget:  uint8(pb.DifficultyTarget), // silently removes invalid values
		Nonce:             pb.Nonce,
	}, nil
}

func toInvVectors(inventoryPb []*pb.InvVector) ([]*inv.InvVector, error) {
	inventory := make([]*inv.InvVector, len(inventoryPb))
	for i, v := range inventoryPb {
		domainVec, err := toInvVector(v)
		if err != nil {
			return nil, err
		}
		inventory[i] = &domainVec
	}
	return inventory, nil
}

func toTransactions(transactions []*pb.Transaction) ([]transaction.Transaction, error) {
	if transactions == nil {
		return nil, fmt.Errorf("transactions must not be nil")
	}
	txs := make([]transaction.Transaction, len(transactions))
	for i, tx := range transactions {
		t, err := toTx(tx)
		if err != nil {
			return nil, err
		}
		txs[i] = t
	}

	return txs, nil
}

func toTx(tx *pb.Transaction) (transaction.Transaction, error) {
	if tx == nil {
		return transaction.Transaction{}, fmt.Errorf("transaction must not be nil")
	}
	if tx.Inputs == nil {
		return transaction.Transaction{}, fmt.Errorf("transaction inputs must not be nil")
	}
	if tx.Outputs == nil {
		return transaction.Transaction{}, fmt.Errorf("transaction outputs must not be nil")
	}

	inputs, err := toInputs(tx.Inputs)
	if err != nil {
		return transaction.Transaction{}, err
	}
	outputs, err := toOutputs(tx.Outputs)
	if err != nil {
		return transaction.Transaction{}, err
	}

	return transaction.Transaction{
		Inputs:  inputs,
		Outputs: outputs,
	}, nil
}

func toInputs(inPb []*pb.TxInput) ([]transaction.Input, error) {
	if inPb == nil {
		return nil, fmt.Errorf("inputs must not be nil")
	}

	inputs := make([]transaction.Input, len(inPb))
	for i, in := range inPb {
		input, err := toInput(in)
		if err != nil {
			return nil, err
		}
		inputs[i] = input
	}

	return inputs, nil
}

func toInput(in *pb.TxInput) (transaction.Input, error) {
	if in == nil {
		return transaction.Input{}, fmt.Errorf("input must not be nil")
	}

	var prevTxId transaction.TransactionID
	copy(prevTxId[:], in.PrevTxHash)

	var pubKey transaction.PubKey
	copy(pubKey[:], in.PublicKey)

	return transaction.Input{
		PrevTxID:    prevTxId,
		OutputIndex: in.OutputIndex,
		Signature:   in.Signature,
		PubKey:      pubKey,
	}, nil
}

func toOutputs(outPb []*pb.TxOutput) ([]transaction.Output, error) {
	if outPb == nil {
		return nil, fmt.Errorf("outputs must not be nil")
	}

	outputs := make([]transaction.Output, len(outPb))
	for i, out := range outPb {
		output, err := toOutput(out)
		if err != nil {
			return nil, err
		}
		outputs[i] = output
	}

	return outputs, nil
}

func toOutput(outPb *pb.TxOutput) (transaction.Output, error) {
	if outPb == nil {
		return transaction.Output{}, fmt.Errorf("output must not be nil")
	}
	var pubKeyHash transaction.PubKeyHash
	copy(pubKeyHash[:], outPb.PublicKeyHash)

	return transaction.Output{
		Value:      outPb.Value,
		PubKeyHash: pubKeyHash,
	}, nil
}

func toMerkleProofs(proofsPb []*pb.MerkleProof) ([]block.MerkleProof, error) {
	if proofsPb == nil {
		return nil, fmt.Errorf("proofs must not be nil")
	}

	proofs := make([]block.MerkleProof, len(proofsPb))
	for i, p := range proofsPb {
		if p == nil {
			return nil, fmt.Errorf("proofs[%d] must not be nil", i)
		}

		proof, err := toMerkleProof(p)
		if err != nil {
			return nil, err
		}

		proofs[i] = proof
	}

	return proofs, nil
}

func toMerkleProof(pb *pb.MerkleProof) (block.MerkleProof, error) {
	if pb == nil {
		return block.MerkleProof{}, fmt.Errorf("merkle proof must not be nil")
	}
	if pb.Siblings == nil {
		return block.MerkleProof{}, fmt.Errorf("merkle proof siblings must not be nil")
	}

	tx, err := toTx(pb.Transaction)
	if err != nil {
		return block.MerkleProof{}, err
	}

	siblings := make([]common.Hash, len(pb.Siblings))
	for i, s := range pb.Siblings {
		if len(s) != common.HashSize {
			return block.MerkleProof{}, fmt.Errorf("merkle proof sibling[%d] must be %d bytes long", i, common.HashSize)
		}

		var hash common.Hash
		copy(hash[:], s[:])
		siblings[i] = hash
	}

	return block.MerkleProof{
		Transaction: tx,
		Siblings:    siblings,
		Index:       pb.Index,
	}, nil
}
