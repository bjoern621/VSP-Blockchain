package api_wallet

import "s3b/vsp-blockchain/p2p-blockchain/wallet/core/keys"

// KeyGeneratorApi is the external API for generating and getting keysets.
type KeyGeneratorApi interface {
	// GenerateKeyset Generates a completely new Keyset with a new random private Key
	GenerateKeyset() keys.Keyset

	// GetKeyset Gets the complete keyset from the raw private key
	GetKeyset(privateKey [32]byte) keys.Keyset

	// GetKeysetFromWIF Gets the complete keyset from the WIF encoded private key
	GetKeysetFromWIF(privateKeyWIF string) keys.Keyset
}

type KeyGeneratorApiImpl struct {
	keyGenerator keys.KeyGenerator
}

func NewKeyGeneratorApiImpl(keyGenerator keys.KeyGenerator) *KeyGeneratorApiImpl {
	return &KeyGeneratorApiImpl{
		keyGenerator: keyGenerator,
	}
}

func (k KeyGeneratorApiImpl) GenerateKeyset() keys.Keyset {
	return k.keyGenerator.GenerateKeyset()
}

func (k KeyGeneratorApiImpl) GetKeyset(privateKey [32]byte) keys.Keyset {
	return k.keyGenerator.GetKeyset(privateKey)
}

func (k KeyGeneratorApiImpl) GetKeysetFromWIF(privateKeyWIF string) keys.Keyset {
	return k.keyGenerator.GetKeysetFromWIF(privateKeyWIF)
}
