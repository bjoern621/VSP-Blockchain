package transaction

type Output struct {
	// Value is the amount of cryptocurrency in smallest units.
	// The smallest unit is 1 V$Goin. This means fractional amounts of V$Goins are not allowed.
	Value uint64
	// PubKeyHash is the 20 bytes hash of the public key that can spend this output.
	// Also known as the "vs address".
	PubKeyHash PubKeyHash
}

func (out *Output) Clone() Output {
	return Output{
		Value:      out.Value,
		PubKeyHash: out.PubKeyHash,
	}
}
