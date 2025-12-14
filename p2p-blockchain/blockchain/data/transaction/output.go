package transaction

import "s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api/blockchain/dto"

type Output struct {
	Value      uint64
	PubKeyHash PubKeyHash
}

func NewOutputFromDTO(o dto.TxOutputDTO) Output {
	return Output{
		Value: o.Value,
		//TODO: Was passiert hier? TxOutput hat im Proto nur PublicKeyScript, braucht hier aber PubKeyHash??? PubKeyHash: o.PublicKeyScript
	}
}

func (out *Output) Clone() Output {
	return Output{
		Value:      out.Value,
		PubKeyHash: out.PubKeyHash,
	}
}
