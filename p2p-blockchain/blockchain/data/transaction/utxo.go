package transaction

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/constants"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api/blockchain/dto"
	"sort"
)

type UTXO struct {
	TxID        TransactionID
	OutputIndex uint32
	Output      Output
}
type TransactionID [constants.HashSize]byte

func NewTransactionIDFromDTO(h dto.Hash) TransactionID {
	var id TransactionID
	copy(id[:], h[:])
	return id
}

func selectUTXOs(utxos []UTXO, amount uint64) (selected []UTXO, total uint64) {
	sort.Slice(utxos, func(i, j int) bool {
		return utxos[i].Output.Value < utxos[j].Output.Value
	})

	for _, u := range utxos {
		selected = append(selected, u)
		total += u.Output.Value
		if total >= amount {
			break
		}
	}
	return
}
