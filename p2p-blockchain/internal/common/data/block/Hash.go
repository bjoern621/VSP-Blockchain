package block

import (
	"fmt"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
)

type Hash [common.HashSize]byte

func NewHash(bytes []byte) (Hash, error) {
	if (len(bytes)) != common.HashSize {
		return Hash{}, fmt.Errorf("invalid hash length")
	}
	var hash Hash
	copy(hash[:], bytes)
	return hash, nil
}
