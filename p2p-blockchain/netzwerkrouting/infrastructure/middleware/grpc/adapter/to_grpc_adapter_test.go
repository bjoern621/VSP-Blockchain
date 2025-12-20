package adapter

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToGrpcGetDataMsg(t *testing.T) {
	t.Run("Valid inventory", func(t *testing.T) {
		hash := common.Hash{1, 2, 3}
		inventory := []*block.InvVector{
			{InvType: block.InvType(pb.InvType_MSG_TX), Hash: hash},
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
		inventory := []*block.InvVector{
			{InvType: block.InvType(pb.InvType_MSG_BLOCK), Hash: common.Hash{4, 5, 6}},
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
