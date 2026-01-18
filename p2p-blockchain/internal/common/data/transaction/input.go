package transaction

type Input struct {
	PrevTxID    TransactionID
	OutputIndex uint32
	// Signature is the digital signature proving ownership of the referenced output.
	// The referenced output is the combination of PrevTxID and OutputIndex.
	Signature []byte
	// PubKey is the public key corresponding to the private key that created the signature.
	// This is also the public key that corresponds to the public key hash in the referenced output.
	// It is included here to allow verification of the signature as the public key hash in the referenced output can't be used directly.
	//
	// (In Bitcoin, this is not needed as the public key is included in the scriptSig of the input.)
	PubKey PubKey
}

func (in *Input) Clone() Input {
	return Input{
		PrevTxID:    in.PrevTxID,
		OutputIndex: in.OutputIndex,
		Signature:   append([]byte(nil), in.Signature...), //deep copy
		PubKey:      in.PubKey,
	}
}
