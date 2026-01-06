package transaction

import (
	"crypto/sha256"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"

	"github.com/btcsuite/btcd/btcec/v2"
)

type PubKeyHash [common.PublicKeyHashSize]byte

type PubKey [common.PublicKeySize]byte

type PrivateKey [common.PrivateKeySize]byte

// Hash160 Uses SHA256 to double Hash the 33 Byte public key to a 20 Byte public key hash also known as Address
func Hash160(pub PubKey) PubKeyHash {
	sha := sha256.Sum256(pub[:])

	double := sha256.Sum256(sha[:])

	var h PubKeyHash
	copy(h[:], double[:20])
	return h
}

// pubFromPriv derives the compressed public key from the given private key
func pubFromPriv(privateKey PrivateKey) PubKey {
	_, publicKey := btcec.PrivKeyFromBytes(privateKey[:])
	compressedPubKey := publicKey.SerializeCompressed()

	var pubKey PubKey
	copy(pubKey[:], compressedPubKey)
	return pubKey
}
