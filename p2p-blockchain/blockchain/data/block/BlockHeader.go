package block

import "s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api/blockchain/dto"

// BlockHeader represents the header of a block in the blockchain.
type BlockHeader struct {
	Hash             Hash
	MerkleRoot       Hash
	Timestamp        int64
	DifficultyTarget uint32
	Nonce            uint32
}

func NewBlockHeaderFromDTO(h dto.BlockHeaderDTO) BlockHeader {
	return BlockHeader{
		Hash:             NewHashFromDTO(h.PrevBlockHash),
		MerkleRoot:       NewHashFromDTO(h.MerkleRoot),
		Timestamp:        h.Timestamp,
		DifficultyTarget: h.DifficultyTarget,
		Nonce:            h.Nonce,
	}
}

func NewBlockHeadersFromDTO(m dto.BlockHeadersDTO) []BlockHeader {
	headers := make([]BlockHeader, 0, len(m.Headers))
	for i := range m.Headers {
		headers = append(headers, NewBlockHeaderFromDTO(m.Headers[i]))
	}
	return headers
}
