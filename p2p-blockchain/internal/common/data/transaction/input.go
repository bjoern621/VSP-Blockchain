package transaction

type Input struct {
	PrevTxID    TransactionID
	OutputIndex uint32
	Signature   []byte
	PubKey      PubKey
	Sequence    uint32
}

func (in *Input) Clone() Input {
	return Input{
		PrevTxID:    in.PrevTxID,
		OutputIndex: in.OutputIndex,
		Signature:   append([]byte(nil), in.Signature...), //deep copy
		PubKey:      in.PubKey,
		Sequence:    in.Sequence,
	}
}
