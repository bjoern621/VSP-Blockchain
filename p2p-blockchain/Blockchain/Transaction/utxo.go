package Transaction

import "sort"

type UTXO struct {
	TxID        TransactionID
	OutputIndex uint32
	Output      Output
}
type TransactionID [32]byte

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
