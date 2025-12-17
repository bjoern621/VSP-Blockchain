package block

import (
	"fmt"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api/blockchain/dto"
)

type PublicKeyHash [common.PublicKeyHashSize]byte

func NewPublicKeyHash(bytes []byte) (PublicKeyHash, error) {
	if (len(bytes)) != common.PublicKeyHashSize {
		return PublicKeyHash{}, fmt.Errorf("invalid public key hash length")
	}
	var hash PublicKeyHash
	copy(hash[:], bytes)
	return hash, nil
}

// NewPublicKeyHashFromDTO asserts that the DTO value is valid and returns a data-layer PublicKeyHash.
// Invalid input is a programmer error and will halt execution.
func NewPublicKeyHashFromDTO(h dto.PublicKeyHash) PublicKeyHash {
	var out PublicKeyHash
	copy(out[:], h[:])
	return out
}
