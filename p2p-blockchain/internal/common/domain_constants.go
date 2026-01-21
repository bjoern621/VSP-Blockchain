package common

const (
	PrivateKeySize    = 32
	PublicKeyHashSize = 20
	PublicKeySize     = 33
	// HashSize defines the size of (1) block-, (2) transaction- and (3) merkle root-hashes in bytes.
	HashSize                                             = 32
	TransactionFee                                uint64 = 1
	TransactionBlockHeightDifferenceForAcceptance        = 1
)
