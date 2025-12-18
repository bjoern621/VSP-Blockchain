package api

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/wallet/core/keys"
	"testing"
)

// Private key Hex: 904c4260d266de0848bb7fba161242fe37b8982c3cdc552564dfebc4fd42a527
// Private Key WIF: 5JuqUeBRXKFXqetT6LujZGaBR9hfWCtCq7iSiFYeaBfp6PBJu2A
// Public Key Hex: 04896a441587cb491a2ab9b2aee8a39202b4c79982f6f97009c2fbf5228c631c88a0527eeee6a6dd574347ffaad90cc327119ac6d2a39440284530d5a4c9911f8c
// Address: 1BLRMuPN13ouhRWhnKCjth1ZcGYfeJHEzR
var sampleKeyset = common.Keyset{
	PrivateKey:    [32]byte{0x90, 0x4c, 0x42, 0x60, 0xd2, 0x66, 0xde, 0x08, 0x48, 0xbb, 0x7f, 0xba, 0x16, 0x12, 0x42, 0xfe, 0x37, 0xb8, 0x98, 0x2c, 0x3c, 0xdc, 0x55, 0x25, 0x64, 0xdf, 0xeb, 0xc4, 0xfd, 0x42, 0xa5, 0x27},
	PrivateKeyWif: "5JuqUeBRXKFXqetT6LujZGaBR9hfWCtCq7iSiFYeaBfp6PBJu2A",
	PublicKey:     [65]byte{0x04, 0x89, 0x6a, 0x44, 0x15, 0x87, 0xcb, 0x49, 0x1a, 0x2a, 0xb9, 0xb2, 0xae, 0xe8, 0xa3, 0x92, 0x02, 0xb4, 0xc7, 0x99, 0x82, 0xf6, 0xf9, 0x70, 0x09, 0xc2, 0xfb, 0xf5, 0x22, 0x8c, 0x63, 0x1c, 0x88, 0xa0, 0x52, 0x7e, 0xee, 0xe6, 0xa6, 0xdd, 0x57, 0x43, 0x47, 0xff, 0xaa, 0xd9, 0x0c, 0xc3, 0x27, 0x11, 0x9a, 0xc6, 0xd2, 0xa3, 0x94, 0x40, 0x28, 0x45, 0x30, 0xd5, 0xa4, 0xc9, 0x91, 0x1f, 0x8c},
	VSAddress:     "1BLRMuPN13ouhRWhnKCjth1ZcGYfeJHEzR",
}

func SetupKeyGeneratorApiImpl() *KeyGeneratorApiImpl {
	keyEncodingsImpl := keys.NewKeyEncodingsImpl()
	keyGeneratorImpl := keys.NewKeyGeneratorImpl(keyEncodingsImpl, keyEncodingsImpl)
	keyGeneratorApiImpl := NewKeyGeneratorApiImpl(keyGeneratorImpl)
	return keyGeneratorApiImpl
}

func TestKeyGeneratorApiImpl_GenerateKeyset(t *testing.T) {
	keyGeneratorApiImpl := SetupKeyGeneratorApiImpl()

	keyset := keyGeneratorApiImpl.GenerateKeyset()
	if len(keyset.PrivateKeyWif) != 51 {
		t.Errorf("keyset.PrivateKeyWif should be 51 bytes")
	}
}

func TestKeyGeneratorApiImpl_GetKeyset(t *testing.T) {
	keyGeneratorApiImpl := SetupKeyGeneratorApiImpl()

	keyset := keyGeneratorApiImpl.GetKeyset(sampleKeyset.PrivateKey)

	if keyset.PrivateKeyWif != sampleKeyset.PrivateKeyWif {
		t.Errorf("keyset.PrivateKeyWif does not match the given private key")
	}
}

func TestKeyGeneratorApiImpl_GetKeysetFromWIF(t *testing.T) {
	keyGeneratorApiImpl := SetupKeyGeneratorApiImpl()

	keyset, err := keyGeneratorApiImpl.GetKeysetFromWIF(sampleKeyset.PrivateKeyWif)

	if err != nil {
		t.Errorf("error return from GetKeysetFromWIF")
	} else if keyset.PrivateKey != sampleKeyset.PrivateKey {
		t.Errorf("keyset.PrivateKey does not match the given private key")
	}
}
