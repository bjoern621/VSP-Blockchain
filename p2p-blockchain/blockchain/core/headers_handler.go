package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/inv"

	"bjoernblessin.de/go-utils/util/logger"
)

func (b *Blockchain) Headers(blockHeaders []*block.BlockHeader, peerID common.PeerId) {
	if !b.CheckPeerIsConnected(peerID) {
		return
	}

	headerHashes := make([]common.Hash, len(blockHeaders))
	for i, header := range blockHeaders {
		headerHashes[i] = header.Hash()
	}
	logger.Infof("[headers_handler] Headers Message received: %d headers from %v: %v", len(blockHeaders), peerID, headerHashes)

	if len(blockHeaders) == 0 {
		return
	}

	unknownValidHeaders := make([]*inv.InvVector, 0)

	for i, header := range blockHeaders {
		if ok, err := b.blockValidator.ValidateHeaderOnly(*header); !ok {
			logger.Warnf("[headers_handler] Invalid header at index %d from %v: %v", i, peerID, err)
			continue
		}

		headerHash := header.Hash()
		if _, err := b.blockStore.GetBlockByHash(headerHash); err == nil {
			continue
		}

		unknownValidHeaders = append(unknownValidHeaders, &inv.InvVector{
			InvType: inv.InvTypeMsgBlock,
			Hash:    headerHash,
		})
	}

	if len(unknownValidHeaders) > 0 {
		logger.Infof("[headers_handler] Requesting %d unknown blocks from %v", len(unknownValidHeaders), peerID)
		b.blockchainMsgSender.SendGetData(unknownValidHeaders, peerID)
	}
}
