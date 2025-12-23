package api

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/wallet/core/keys"
	"testing"
)

// Private key Hex: 904c4260d266de0848bb7fba161242fe37b8982c3cdc552564dfebc4fd42a527
// Private Key WIF: 5JuqUeBRXKFXqetT6LujZGaBR9hfWCtCq7iSiFYeaBfp6PBJu2A
// Public Key Hex: 02896a441587cb491a2ab9b2aee8a39202b4c79982f6f97009c2fbf5228c631c88
// Address: 1BLRMuPN13ouhRWhnKCjth1ZcGYfeJHEzR
var sampleKeyset1 = common.Keyset{
	PrivateKey:    [common.PrivateKeySize]byte{0x90, 0x4c, 0x42, 0x60, 0xd2, 0x66, 0xde, 0x08, 0x48, 0xbb, 0x7f, 0xba, 0x16, 0x12, 0x42, 0xfe, 0x37, 0xb8, 0x98, 0x2c, 0x3c, 0xdc, 0x55, 0x25, 0x64, 0xdf, 0xeb, 0xc4, 0xfd, 0x42, 0xa5, 0x27},
	PrivateKeyWif: "5JuqUeBRXKFXqetT6LujZGaBR9hfWCtCq7iSiFYeaBfp6PBJu2A",
	PublicKey:     [common.PublicKeySize]byte{0x02, 0x89, 0x6a, 0x44, 0x15, 0x87, 0xcb, 0x49, 0x1a, 0x2a, 0xb9, 0xb2, 0xae, 0xe8, 0xa3, 0x92, 0x02, 0xb4, 0xc7, 0x99, 0x82, 0xf6, 0xf9, 0x70, 0x09, 0xc2, 0xfb, 0xf5, 0x22, 0x8c, 0x63, 0x1c, 0x88},
	VSAddress:     "1BLRMuPN13ouhRWhnKCjth1ZcGYfeJHEzR",
}

// Private key Hex: 4f3bbfa2cc5ef348b95ef85c805a39d585e9dfe4a701d726bc2f941f92c115ad
// Private Key WIF: 5JRBVVyS8Xb4vJA2he7BJ7P6NcGnup5YkHA9sGKF6pdhyRLJb5L
// Public Key Hex: 0339a5088f2a067087b3fa32c1726a078997de310159b1f228f0ec7f44a0a5f2e4
// Address: 1Fy1MtFDdyzWGbveGXBrHSL2WT4SeU2GGc
var sampleKeyset2 = common.Keyset{
	PrivateKey:    [common.PrivateKeySize]byte{0x4f, 0x3b, 0xbf, 0xa2, 0xcc, 0x5e, 0xf3, 0x48, 0xb9, 0x5e, 0xf8, 0x5c, 0x80, 0x5a, 0x39, 0xd5, 0x85, 0xe9, 0xdf, 0xe4, 0xa7, 0x01, 0xd7, 0x26, 0xbc, 0x2f, 0x94, 0x1f, 0x92, 0xc1, 0x15, 0xad},
	PrivateKeyWif: "5JRBVVyS8Xb4vJA2he7BJ7P6NcGnup5YkHA9sGKF6pdhyRLJb5L",
	PublicKey:     [common.PublicKeySize]byte{0x03, 0x39, 0xa5, 0x08, 0x8f, 0x2a, 0x06, 0x70, 0x87, 0xb3, 0xfa, 0x32, 0xc1, 0x72, 0x6a, 0x07, 0x89, 0x97, 0xde, 0x31, 0x01, 0x59, 0xb1, 0xf2, 0x28, 0xf0, 0xec, 0x7f, 0x44, 0xa0, 0xa5, 0xf2, 0xe4},
	VSAddress:     "1Fy1MtFDdyzWGbveGXBrHSL2WT4SeU2GGc",
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

func TestKeyGeneratorApiImpl_GetKeyset_even(t *testing.T) {
	keyGeneratorApiImpl := SetupKeyGeneratorApiImpl()

	keyset := keyGeneratorApiImpl.GetKeyset(sampleKeyset1.PrivateKey)

	if keyset.PrivateKeyWif != sampleKeyset1.PrivateKeyWif {
		t.Errorf("keyset.PrivateKeyWif does not match the given private key")
	}
	if keyset.PublicKey != sampleKeyset1.PublicKey {
		t.Errorf("keyset.PublicKey does not match the given public key")
	}
}

func TestKeyGeneratorApiImpl_GetKeyset_odd(t *testing.T) {
	keyGeneratorApiImpl := SetupKeyGeneratorApiImpl()

	keyset := keyGeneratorApiImpl.GetKeyset(sampleKeyset2.PrivateKey)

	if keyset.PrivateKeyWif != sampleKeyset2.PrivateKeyWif {
		t.Errorf("keyset.PrivateKeyWif does not match the given private key")
	}
	if keyset.PublicKey != sampleKeyset2.PublicKey {
		t.Errorf("keyset.PublicKey does not match the given public key")
	}
}

func TestKeyGeneratorApiImpl_GetKeysetFromWIF(t *testing.T) {
	keyGeneratorApiImpl := SetupKeyGeneratorApiImpl()

	keyset, err := keyGeneratorApiImpl.GetKeysetFromWIF(sampleKeyset1.PrivateKeyWif)

	if err != nil {
		t.Errorf("error return from GetKeysetFromWIF")
	}
	if keyset.PrivateKey != sampleKeyset1.PrivateKey {
		t.Errorf("keyset.PrivateKey does not match the given private key")
	}
	if keyset.PublicKey != sampleKeyset1.PublicKey {
		t.Errorf("keyset.PublicKey does not match the given public key")
	}
}
