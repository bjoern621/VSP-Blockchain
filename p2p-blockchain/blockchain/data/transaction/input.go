package transaction

import "s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api/blockchain/dto"

type Input struct {
	PrevTxID    TransactionID
	OutputIndex uint32
	Signature   []byte
	PubKey      PubKey
	Sequence    uint32
}

func NewInputFromDTO(in dto.TxInputDTO) Input {
	return Input{
		PrevTxID:    NewTransactionIDFromDTO(in.PrevTxHash),
		OutputIndex: in.OutputIndex,
		Signature:   in.Signature,
		PubKey:      PubKey(in.PublicKey),
		Sequence:    in.Sequence,
	}
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
