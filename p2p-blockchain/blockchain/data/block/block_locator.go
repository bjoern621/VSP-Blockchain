package block

import "s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api/blockchain/dto"

type BlockLocator struct {
	BlockLocatorHashes []Hash
	StopHash           Hash
}

func NewBlockLocatorFromDTO(m dto.BlockLocatorDTO) BlockLocator {
	hashes := make([]Hash, 0, len(m.BlockLocatorHashes))
	for i := range m.BlockLocatorHashes {
		hashes = append(hashes, NewHashFromDTO(m.BlockLocatorHashes[i]))
	}
	return BlockLocator{
		BlockLocatorHashes: hashes,
		StopHash:           NewHashFromDTO(m.HashStop),
	}
}
