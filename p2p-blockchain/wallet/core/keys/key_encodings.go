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
	WifToPrivateKey(wif string) [32]byte
	Base58CheckToBytes(input string) ([]byte, byte)
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

func (keyEncodings *KeyEncodingsImpl) WifToPrivateKey(wif string) [32]byte {
	bytes, version := keyEncodings.Base58CheckToBytes(wif)

	if version != 0x80 || len(bytes) != 32 {
		//TODO fehlerbehandlung
		fmt.Println("Ist kein WIF!")
	}

	return [32]byte(bytes)
}

func (keyEncodings *KeyEncodingsImpl) Base58CheckToBytes(input string) ([]byte, byte) {
	bytes, err := base58.Decode(input)
	if err != nil {
		//TODO Fehlerbehandlung
		//logigng fehler werfen
		fmt.Println("Error decoding base58")
	}
	version := bytes[0]
	payload := bytes[1 : len(bytes)-4]
	checksumBytes := bytes[len(bytes)-4:]

	if !bt.Equal(checksumBytes, keyEncodings.getFirstFourChecksumBytes([]byte{version}, payload)) {
		//TODO Fehlerbehandlung
		fmt.Println("base58 check failed")
	}
	return payload, version
}
