package Transaction

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"
)

type Transaction struct {
	Inputs   []Input
	Outputs  []Output
	LockTime uint64
}

func NewTransaction(
	utxos []UTXO,
	toPubKeyHash PubKeyHash,
	amount uint64,
	fee uint64,
	privateKey *ecdsa.PrivateKey,
) (*Transaction, error) {

	selected, total := selectUTXOs(utxos, amount+fee)
	if total < amount+fee {
		return nil, fmt.Errorf("insufficient funds")
	}
	change := total - amount - fee

	tx := fillInTransactionData(selected, amount, toPubKeyHash, privateKey, change)

	if err := tx.Sign(privateKey, selected); err != nil {
		return nil, err
	}

	return tx, nil
}

func fillInTransactionData(selected []UTXO, amount uint64, toPubKeyHash PubKeyHash, privateKey *ecdsa.PrivateKey, change uint64) *Transaction {
	tx := &Transaction{
		LockTime: 0,
	}

	tx.addUnsignedInputs(selected)

	tx.addOutput(amount, toPubKeyHash)
	if change > 0 {
		tx.addChange(change, privateKey)
	}
	return tx
}

func (tx *Transaction) addChange(change uint64, privateKey *ecdsa.PrivateKey) {
	compressedPubKey := elliptic.MarshalCompressed(privateKey.Curve, privateKey.X, privateKey.Y)
	var pubKey PubKey
	copy(pubKey[:], compressedPubKey)
	ownAddress := Hash160(pubKey)

	tx.Outputs = append(tx.Outputs, Output{
		Value:      change,
		PubKeyHash: ownAddress,
	})
}

func (tx *Transaction) addOutput(amount uint64, toPubKeyHash PubKeyHash) {
	tx.Outputs = append(tx.Outputs, Output{
		Value:      amount,
		PubKeyHash: toPubKeyHash,
	})
}

func (tx *Transaction) addUnsignedInputs(selected []UTXO) {
	for _, u := range selected {
		tx.addInput(u)
	}
}

func (tx *Transaction) addInput(u UTXO) {
	tx.Inputs = append(tx.Inputs, Input{
		PrevTxID:    u.TxID,
		OutputIndex: u.OutputIndex,
		Sequence:    0xffffffff,
	})
}

func (tx *Transaction) Clone() *Transaction {
	clone := &Transaction{
		LockTime: tx.LockTime,
	}

	// Deep copy inputs
	clone.Inputs = make([]Input, len(tx.Inputs))
	for i, in := range tx.Inputs {
		clone.Inputs[i] = in.Clone()
	}

	// Deep copy outputs
	clone.Outputs = make([]Output, len(tx.Outputs))
	for i, out := range tx.Outputs {
		clone.Outputs[i] = out.Clone()
	}

	return clone
}
