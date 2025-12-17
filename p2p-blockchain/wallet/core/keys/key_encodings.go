package keys

import (
	"crypto/sha256"

	"github.com/akamensky/base58"
)

// KeyEncoder Encodes keys to the most common formats
type KeyEncoder interface {
	PrivateKeyToWif(privateKey [32]byte) string
	BytesToBase58Check(bytes []byte, version byte) string
}

// KeyDecoder Decodes keys from the most common formats
type KeyDecoder interface {
}

// KeyEncodingsImpl Implements KeyEncoder and KeyDecoder
type KeyEncodingsImpl struct {
}

func NewKeyEncodingsImpl() *KeyEncodingsImpl {
	return &KeyEncodingsImpl{}
}

func (keyEncodings *KeyEncodingsImpl) PrivateKeyToWif(privateKey [32]byte) string {
	return keyEncodings.BytesToBase58Check(privateKey[:], 0x80)
}

func (keyEncodings *KeyEncodingsImpl) BytesToBase58Check(bytes []byte, version byte) string {

	together := make([]byte, 0, 1+len(bytes)+4)

	//1. part: version
	together = append(together, version)

	//2. part: payload
	together = append(together, bytes...)

	//3. part: checksum
	together = append(together, keyEncodings.getFirstFourChecksumBytes([]byte{version}, bytes)...)

	return base58.Encode(together)
}

func (keyEncodings *KeyEncodingsImpl) getFirstFourChecksumBytes(bytes ...[]byte) []byte {
	h := sha256.New()
	for _, bytes := range bytes {
		h.Write(bytes)
	}
	firstHash := h.Sum(nil)
	secondHash := sha256.Sum256(firstHash)

	return secondHash[:4]
}
