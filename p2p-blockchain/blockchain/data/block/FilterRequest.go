package block

import "s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api/blockchain/dto"

type SetFilterRequest struct {
	PublicKeyHashes []PublicKeyHash
}

func NewSetFilterRequestFromDTO(m dto.SetFilterRequestDTO) SetFilterRequest {
	hashes := make([]PublicKeyHash, 0, len(m.PublicKeyHashes))
	for i := range m.PublicKeyHashes {
		hashes = append(hashes, NewPublicKeyHashFromDTO(m.PublicKeyHashes[i]))
	}
	return SetFilterRequest{PublicKeyHashes: hashes}
}
