package common

type Keyset struct {
	PrivateKey    [PrivateKeySize]byte
	PrivateKeyWif string
	PublicKey     [PublicKeySize]byte
	VSAddress     string
}
