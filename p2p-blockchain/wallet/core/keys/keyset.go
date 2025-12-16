package keys

type Keyset struct {
	PrivateKey    [32]byte
	PrivateKeyWif string
	PublicKey     [65]byte
	VSAddress     string
}
