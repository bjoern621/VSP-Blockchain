package transaction

import (
	"crypto/sha256"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/constants"
)

type PubKeyHash [constants.PublicKeyHashSize]byte

type PubKey [constants.PublicKeySize]byte

// Hash160 Uses SHA256 to double Hash the 33 Byte public key to a 20 Byte public key hash also known as Address
func Hash160(pub PubKey) PubKeyHash {
	sha := sha256.Sum256(pub[:])

	double := sha256.Sum256(sha[:])

	var h PubKeyHash
	copy(h[:], double[:20])
	return h
}
