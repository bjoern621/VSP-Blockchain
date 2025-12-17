package api

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api/blockchain/dto"
)

type BlockchainService interface {
	SendGetData(dto dto.GetDataMsgDTO, peerId common.PeerId)
}
