package common

import "encoding/hex"

type Hash [HashSize]byte

func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}
