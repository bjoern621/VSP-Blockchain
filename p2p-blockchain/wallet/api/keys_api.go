package api

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/wallet/core/keys"
)

// KeyGeneratorApi is the external API for generating and getting keysets.
// Part of WalletAppAPI.
type KeyGeneratorApi interface {
	// GenerateKeyset Generates a completely new Keyset with a new random private Key
	GenerateKeyset() common.Keyset

	// GetKeyset Gets the complete keyset from the raw private key
	GetKeyset(privateKey [common.PrivateKeySize]byte) common.Keyset

	// GetKeysetFromWIF Gets the complete keyset from the WIF encoded private key
	GetKeysetFromWIF(privateKeyWIF string) (common.Keyset, error)
}

type KeyGeneratorApiImpl struct {
	keyGenerator keys.KeyGenerator
}

func NewKeyGeneratorApiImpl(keyGenerator keys.KeyGenerator) *KeyGeneratorApiImpl {
	return &KeyGeneratorApiImpl{
		keyGenerator: keyGenerator,
	}
}

func (k *KeyGeneratorApiImpl) GenerateKeyset() common.Keyset {
	return k.keyGenerator.GenerateKeyset()
}

func (k *KeyGeneratorApiImpl) GetKeyset(privateKey [common.PrivateKeySize]byte) common.Keyset {
	return k.keyGenerator.GetKeyset(privateKey)
}

func (k *KeyGeneratorApiImpl) GetKeysetFromWIF(privateKeyWIF string) (common.Keyset, error) {
	return k.keyGenerator.GetKeysetFromWIF(privateKeyWIF)
}
