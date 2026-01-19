package adapter

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/inv"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToGrpcGetDataMsg(t *testing.T) {
	t.Run("Valid inventory", func(t *testing.T) {
		hash := common.Hash{1, 2, 3}
		inventory := []*inv.InvVector{
			{InvType: inv.InvType(pb.InvType_MSG_TX), Hash: hash},
		}

		msg, err := ToGrpcGetDataMsg(inventory)
		assert.NoError(t, err)
		assert.NotNil(t, msg)
		assert.Equal(t, 1, len(msg.Inventory))
		assert.Equal(t, pb.InvType_MSG_TX, msg.Inventory[0].Type)
		assert.Equal(t, hash[:], msg.Inventory[0].Hash)
	})

	t.Run("Nil inventory error", func(t *testing.T) {
		msg, err := ToGrpcGetDataMsg(nil)
		assert.Error(t, err)
		assert.Nil(t, msg)
		assert.Contains(t, err.Error(), "inventory must not be nil")
	})
}

func TestToGrpcGetInvMsg(t *testing.T) {
	t.Run("Valid inventory", func(t *testing.T) {
		inventory := []*inv.InvVector{
			{InvType: inv.InvType(pb.InvType_MSG_BLOCK), Hash: common.Hash{4, 5, 6}},
		}

		msg, err := ToGrpcGetInvMsg(inventory)
		assert.NoError(t, err)
		assert.NotNil(t, msg)
		assert.Equal(t, 1, len(msg.Inventory))
		assert.Equal(t, pb.InvType_MSG_BLOCK, msg.Inventory[0].Type)
	})

	t.Run("Nil inventory error", func(t *testing.T) {
		msg, err := ToGrpcGetInvMsg(nil)
		assert.Error(t, err)
		assert.Nil(t, msg)
	})
}

func TestToGrpcBlockMsg(t *testing.T) {
	t.Run("Valid block", func(t *testing.T) {
		prevBlockHash := common.Hash{1, 2, 3, 4}
		merkleRoot := common.Hash{5, 6, 7, 8}

		tx1 := transaction.Transaction{
			Inputs: []transaction.Input{
				{
					PrevTxID:    transaction.TransactionID{9, 10, 11, 12},
					OutputIndex: 0,
					Signature:   []byte{0x01, 0x02},
					PubKey:      transaction.PubKey{0x03, 0x04},
				},
			},
			Outputs: []transaction.Output{
				{
					Value:      100,
					PubKeyHash: transaction.PubKeyHash{13, 14, 15, 16},
				},
			},
		}

		b := &block.Block{
			Header: block.BlockHeader{
				PreviousBlockHash: prevBlockHash,
				MerkleRoot:        merkleRoot,
				Timestamp:         1234567890,
				DifficultyTarget:  28,
				Nonce:             12345,
			},
			Transactions: []transaction.Transaction{tx1},
		}

		msg, err := ToGrpcBlockMsg(b)
		assert.NoError(t, err)
		assert.NotNil(t, msg)
		assert.NotNil(t, msg.Block)

		pbHeader := msg.Block.Header
		assert.Equal(t, prevBlockHash[:], pbHeader.PrevBlockHash)
		assert.Equal(t, merkleRoot[:], pbHeader.MerkleRoot)
		assert.Equal(t, int64(1234567890), pbHeader.Timestamp)
		assert.Equal(t, uint32(28), pbHeader.DifficultyTarget)
		assert.Equal(t, uint32(12345), pbHeader.Nonce)

		assert.Equal(t, 1, len(msg.Block.Transactions))
		pbTx := msg.Block.Transactions[0]
		assert.Equal(t, 1, len(pbTx.Inputs))
		assert.Equal(t, 1, len(pbTx.Outputs))
	})

	t.Run("Nil block error", func(t *testing.T) {
		msg, err := ToGrpcBlockMsg(nil)
		assert.Error(t, err)
		assert.Nil(t, msg)
		assert.Contains(t, err.Error(), "block must not be nil")
	})
}

func TestToGrpcTxMsg(t *testing.T) {
	t.Run("Valid transaction", func(t *testing.T) {
		prevTxID := transaction.TransactionID{1, 2, 3, 4}
		pubKey := transaction.PubKey{0xdd, 0xee, 0xff}
		pubKeyHash := transaction.PubKeyHash{5, 6, 7, 8}

		tx := &transaction.Transaction{
			Inputs: []transaction.Input{
				{
					PrevTxID:    prevTxID,
					OutputIndex: 1,
					Signature:   []byte{0xaa, 0xbb, 0xcc},
					PubKey:      pubKey,
				},
			},
			Outputs: []transaction.Output{
				{
					Value:      500,
					PubKeyHash: pubKeyHash,
				},
			},
		}

		msg, err := ToGrpcTxMsg(tx)
		assert.NoError(t, err)
		assert.NotNil(t, msg)
		assert.NotNil(t, msg.Transaction)

		pbTx := msg.Transaction
		assert.Equal(t, 1, len(pbTx.Inputs))
		assert.Equal(t, 1, len(pbTx.Outputs))

		pbInput := pbTx.Inputs[0]
		assert.Equal(t, prevTxID[:], pbInput.PrevTxHash)
		assert.Equal(t, uint32(1), pbInput.OutputIndex)
		assert.Equal(t, []byte{0xaa, 0xbb, 0xcc}, pbInput.Signature)
		assert.Equal(t, pubKey[:], pbInput.PublicKey)

		pbOutput := pbTx.Outputs[0]
		assert.Equal(t, uint64(500), pbOutput.Value)
		assert.Equal(t, pubKeyHash[:], pbOutput.PublicKeyHash)
	})

	t.Run("Nil transaction error", func(t *testing.T) {
		msg, err := ToGrpcTxMsg(nil)
		assert.Error(t, err)
		assert.Nil(t, msg)
		assert.Contains(t, err.Error(), "transaction must not be nil")
	})
}
