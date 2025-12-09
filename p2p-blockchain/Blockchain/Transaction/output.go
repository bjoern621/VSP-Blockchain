package Transaction

type Output struct {
	Value      uint64
	PubKeyHash PubKeyHash
}

func (out *Output) Clone() Output {
	return Output{
		Value:      out.Value,
		PubKeyHash: out.PubKeyHash,
	}
}
