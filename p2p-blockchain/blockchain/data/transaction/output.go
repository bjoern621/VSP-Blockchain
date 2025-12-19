package transaction

import "s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api/blockchain/dto"

type Output struct {
	Value      uint64
	PubKeyHash PubKeyHash
}

func NewOutputFromDTO(o dto.TxOutputDTO) Output {
	return Output{
		Value:      o.Value,
		PubKeyHash: PubKeyHash(o.PublicKeyHash),
	}
}

func (out *Output) Clone() Output {
	return Output{
		Value:      out.Value,
		PubKeyHash: out.PubKeyHash,
	}
}
