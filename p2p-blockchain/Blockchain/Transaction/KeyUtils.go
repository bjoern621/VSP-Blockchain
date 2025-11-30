package Transaction

import (
	"crypto/sha256"
	"golang.org/x/crypto/ripemd160"
)

type PubKeyHash [20]byte

type PubKey [33]byte

func Hash160(pub PubKey) PubKeyHash {
	sha := sha256.Sum256(pub[:])

	r := ripemd160.New()
	r.Write(sha[:])
	full := r.Sum(nil)

	var h PubKeyHash
	copy(h[:], full) // always 20 bytes
	return h
}
