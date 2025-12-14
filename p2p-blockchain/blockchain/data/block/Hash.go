package block

import (
	"fmt"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/constants"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api/blockchain/dto"
)

type Hash [constants.HashSize]byte

func NewHash(bytes []byte) (Hash, error) {
	if (len(bytes)) != constants.HashSize {
		return Hash{}, fmt.Errorf("invalid hash length")
	}
	var hash Hash
	copy(hash[:], bytes)
	return hash, nil
}

func NewHashFromDTO(h dto.Hash) Hash {
	var out Hash
	copy(out[:], h[:])
	return out
}
