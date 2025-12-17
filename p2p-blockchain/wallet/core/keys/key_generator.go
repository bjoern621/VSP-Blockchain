package keys

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
)

// KeyGenerator Interface for API for creating / getting keysets
type KeyGenerator interface {
	// GenerateKeyset Generates a completely new Keyset with a new random private Key
	GenerateKeyset() common.Keyset

	// GetKeyset Gets the complete keyset from the raw private key
	GetKeyset(privateKey [32]byte) common.Keyset

	// GetKeysetFromWIF Gets the complete keyset from the WIF encoded private key
	GetKeysetFromWIF(privateKeyWIF string) (common.Keyset, error)
}

type KeyGeneratorImpl struct {
	encoder KeyEncoder
	decoder KeyDecoder
}

func NewKeyGeneratorImpl(encoder KeyEncoder, decoder KeyDecoder) *KeyGeneratorImpl {
	return &KeyGeneratorImpl{
		encoder: encoder,
		decoder: decoder,
	}
}

// public functions
// Not fully implemented yet!
func (generator *KeyGeneratorImpl) GenerateKeyset() common.Keyset {
	//private Key generieren
	privateKey := generator.createPrivateKey()
	return generator.GetKeyset(privateKey)
}

// Not fully implemented yet!
func (generator *KeyGeneratorImpl) GetKeyset(privateKey [32]byte) common.Keyset {
	return common.Keyset{
		PrivateKey:    privateKey,
		PrivateKeyWif: generator.encoder.PrivateKeyToWif(privateKey),
	}
}

// Not fully implemented yet!
func (generator *KeyGeneratorImpl) GetKeysetFromWIF(privateKeyWIF string) (common.Keyset, error) {
	privateKey, err := generator.decoder.WifToPrivateKey(privateKeyWIF)
	if err != nil {
		return common.Keyset{}, fmt.Errorf("error decoding WIF: %w", err)
	}
	return common.Keyset{
		PrivateKey:    privateKey,
		PrivateKeyWif: privateKeyWIF,
	}, nil
}

//private functions

// n = 1,158.. *10^77
var nMinusOneBytes = [32]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFE, 0xBA, 0xAE, 0xDC, 0xE6, 0xAF, 0x48, 0xA0, 0x3B, 0xBF, 0xD2, 0x5E, 0x8C, 0xD0, 0x36, 0x41, 0x40}
var nMinusOne = new(big.Int).SetBytes(nMinusOneBytes[:])

func (generator *KeyGeneratorImpl) createPrivateKey() [32]byte {
	for {
		//1. create random 256 bits
		var randomBytes [32]byte
		rand.Read(randomBytes[:]) //nolint:errcheck

		//2. create hash of random bits
		key := sha256.Sum256(randomBytes[:])

		//3. key < n-1, otherwise create e new key
		keyNum := new(big.Int).SetBytes(key[:])
		if keyNum.Cmp(nMinusOne) == -1 {
			return key
		}
	}
}
