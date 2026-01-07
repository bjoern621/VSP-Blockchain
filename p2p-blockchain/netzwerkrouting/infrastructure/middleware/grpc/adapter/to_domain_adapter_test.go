package adapter

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToInvVectorsFromInvMsg(t *testing.T) {
	t.Run("Valid msg", func(t *testing.T) {
		pbMsg := &pb.InvMsg{
			Inventory: []*pb.InvVector{
				{Type: pb.InvType_MSG_TX, Hash: make([]byte, common.HashSize)},
			},
		}
		vecs, err := ToInvVectorsFromInvMsg(pbMsg)
		assert.NoError(t, err)
		assert.Len(t, vecs, 1)
	})

	t.Run("Nil msg", func(t *testing.T) {
		_, err := ToInvVectorsFromInvMsg(nil)
		assert.Error(t, err)
	})

	t.Run("Invalid hash length", func(t *testing.T) {
		pbMsg := &pb.InvMsg{
			Inventory: []*pb.InvVector{
				{Type: pb.InvType_MSG_TX, Hash: []byte{1, 2}}, // Too short
			},
		}
		_, err := ToInvVectorsFromInvMsg(pbMsg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid hash length")
	})
}

func TestToBlockFromBlockMsg(t *testing.T) {
	t.Run("Valid Block", func(t *testing.T) {
		pbMsg := &pb.BlockMsg{
			Block: &pb.Block{
				Header: &pb.BlockHeader{
					PrevBlockHash: make([]byte, common.HashSize),
					MerkleRoot:    make([]byte, common.HashSize),
				},
				Transactions: []*pb.Transaction{
					{
						Inputs:  []*pb.TxInput{{PrevTxHash: make([]byte, common.HashSize)}},
						Outputs: []*pb.TxOutput{{Value: 100, PublicKeyHash: make([]byte, common.HashSize)}},
					},
				},
			},
		}
		b, err := ToBlockFromBlockMsg(pbMsg)
		assert.NoError(t, err)
		assert.Len(t, b.Transactions, 1)
	})

	t.Run("Missing transactions", func(t *testing.T) {
		pbMsg := &pb.BlockMsg{
			Block: &pb.Block{
				Header: &pb.BlockHeader{
					PrevBlockHash: make([]byte, common.HashSize),
					MerkleRoot:    make([]byte, common.HashSize),
				},
				Transactions: nil,
			},
		}
		_, err := ToBlockFromBlockMsg(pbMsg)
		assert.Error(t, err)
	})
}

func TestToBlockLocator(t *testing.T) {
	t.Run("Valid Locator", func(t *testing.T) {
		pbLoc := &pb.BlockLocator{
			BlockLocatorHashes: [][]byte{make([]byte, common.HashSize)},
			HashStop:           make([]byte, common.HashSize),
		}
		loc, err := ToBlockLocator(pbLoc)
		assert.NoError(t, err)
		assert.Len(t, loc.BlockLocatorHashes, 1)
	})

	t.Run("Nil hashes error", func(t *testing.T) {
		pbLoc := &pb.BlockLocator{
			BlockLocatorHashes: nil,
			HashStop:           make([]byte, common.HashSize),
		}
		_, err := ToBlockLocator(pbLoc)
		assert.Error(t, err)
	})
}

func TestToSetFilterRequestFromSetFilterRequest(t *testing.T) {
	t.Run("Valid Request", func(t *testing.T) {
		pbReq := &pb.SetFilterRequest{
			PublicKeyHashes: [][]byte{make([]byte, common.HashSize)},
		}
		req, err := ToSetFilterRequestFromSetFilterRequest(pbReq)
		assert.NoError(t, err)
		assert.Len(t, req.PublicKeyHashes, 1)
	})

	t.Run("Invalid Hash Length", func(t *testing.T) {
		pbReq := &pb.SetFilterRequest{
			PublicKeyHashes: [][]byte{{1, 2, 3}},
		}
		_, err := ToSetFilterRequestFromSetFilterRequest(pbReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bytes long")
	})
}
