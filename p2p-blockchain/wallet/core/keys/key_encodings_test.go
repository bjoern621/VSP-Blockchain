package keys

import (
	bt "bytes"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"testing"
)

// Private key Hex: 904c4260d266de0848bb7fba161242fe37b8982c3cdc552564dfebc4fd42a527
// Private Key WIF: 5JuqUeBRXKFXqetT6LujZGaBR9hfWCtCq7iSiFYeaBfp6PBJu2A
var sampleKeyset = common.Keyset{
	PrivateKey:    [32]byte{0x90, 0x4c, 0x42, 0x60, 0xd2, 0x66, 0xde, 0x08, 0x48, 0xbb, 0x7f, 0xba, 0x16, 0x12, 0x42, 0xfe, 0x37, 0xb8, 0x98, 0x2c, 0x3c, 0xdc, 0x55, 0x25, 0x64, 0xdf, 0xeb, 0xc4, 0xfd, 0x42, 0xa5, 0x27},
	PrivateKeyWif: "5JuqUeBRXKFXqetT6LujZGaBR9hfWCtCq7iSiFYeaBfp6PBJu2A",
}

func TestKeyEncodingsImpl_BytesToBase58Check(t *testing.T) {
	keyEncoder := NewKeyEncodingsImpl()
	str := keyEncoder.BytesToBase58Check(sampleKeyset.PrivateKey[:], 0x80)
	if str != sampleKeyset.PrivateKeyWif {
		t.Errorf("not correct Base58Check output")
	}
}

func TestKeyEncodingsImpl_PrivateKeyToWif(t *testing.T) {
	keyEncoder := NewKeyEncodingsImpl()
	str := keyEncoder.PrivateKeyToWif(sampleKeyset.PrivateKey)
	if str != sampleKeyset.PrivateKeyWif {
		t.Errorf("not correct WIF output")
	}

}

func TestKeyEncodingsImpl_Base58CheckToBytes(t *testing.T) {
	keyDecoder := NewKeyEncodingsImpl()
	resultBytes, version, err := keyDecoder.Base58CheckToBytes(sampleKeyset.PrivateKeyWif)
	if err != nil {
		t.Errorf("unexpected error thrown")
	}
	if version != 0x80 {
		t.Errorf("version not correct")
	}

	if !bt.Equal(resultBytes, sampleKeyset.PrivateKey[:]) {
		t.Errorf("resultBytes not correct")
	}
}

func TestKeyEncodingsImpl_WifToPrivateKey(t *testing.T) {
	keyDecoder := NewKeyEncodingsImpl()
	resultBytes, err := keyDecoder.WifToPrivateKey(sampleKeyset.PrivateKeyWif)
	if err != nil {
		t.Errorf("unexpected error thrown")
	}
	if !bt.Equal(resultBytes[:], sampleKeyset.PrivateKey[:]) {
		t.Errorf("private key not correct")
	}
}
