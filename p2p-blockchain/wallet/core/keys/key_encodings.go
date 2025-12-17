package keys

import (
	"crypto/sha256"
	"fmt"

	bt "bytes"
	"github.com/akamensky/base58"
)

// KeyEncoder Encodes keys to the most common formats
type KeyEncoder interface {
	PrivateKeyToWif(privateKey [32]byte) string
	BytesToBase58Check(bytes []byte, version byte) string
}

// KeyDecoder Decodes keys from the most common formats
type KeyDecoder interface {
	WifToPrivateKey(wif string) ([32]byte, error)
	Base58CheckToBytes(input string) ([]byte, byte, error)
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

func (keyEncodings *KeyEncodingsImpl) WifToPrivateKey(wif string) ([32]byte, error) {
	bytes, version, err := keyEncodings.Base58CheckToBytes(wif)

	if err != nil {
		return [32]byte{}, fmt.Errorf("the wif could not be decoded: %w", err)
	}

	if version != 0x80 || len(bytes) != 32 {
		return [32]byte{}, fmt.Errorf("the Base58Check version byte %x dosent match the required version 0x80 for a WIF", version)
	}

	return [32]byte(bytes), nil
}

func (keyEncodings *KeyEncodingsImpl) Base58CheckToBytes(input string) ([]byte, byte, error) {
	bytes, err := base58.Decode(input)
	if err != nil {
		return []byte{}, 0, fmt.Errorf("error decoding the base58 input (%s): %w", input, err)
	}
	version := bytes[0]
	payload := bytes[1 : len(bytes)-4]
	checksumBytes := bytes[len(bytes)-4:]

	if !bt.Equal(checksumBytes, keyEncodings.getFirstFourChecksumBytes([]byte{version}, payload)) {
		return []byte{}, 0, fmt.Errorf("failed to validate the base58 checksum (from %s) ", input)
	}
	return payload, version, nil
}
