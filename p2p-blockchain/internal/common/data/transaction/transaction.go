package transaction

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"fmt"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
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

// Hash computes the block.Hash similar to the bitcoin wtxid with transaction data including witness
func (tx *Transaction) Hash() common.Hash {
	txId := tx.TransactionId()
	var hash common.Hash
	copy(hash[:], txId[:])
	return hash
}

// TransactionId computes the TransactionID similar to the bitcoin wtxid with transaction data including witness
func (tx *Transaction) TransactionId() TransactionID {
	buf := serializeTransaction(tx)
	hash := doubleSHA256(buf.Bytes())

	var txID TransactionID
	copy(txID[:], hash)
	return txID
}

func serializeTransaction(tx *Transaction) *bytes.Buffer {
	buf := new(bytes.Buffer)
	serializeInputs(tx, buf)
	serializeOutputs(tx, buf)
	writeUint64(buf, tx.LockTime)
	return buf
}

func serializeOutputs(tx *Transaction, buf *bytes.Buffer) {
	writeUint32(buf, uint32(len(tx.Outputs)))
	for _, out := range tx.Outputs {
		writeUint64(buf, out.Value)
		buf.Write(out.PubKeyHash[:])
	}
}

func serializeInputs(tx *Transaction, buf *bytes.Buffer) {
	writeUint32(buf, uint32(len(tx.Inputs)))
	for _, in := range tx.Inputs {
		serializeInput(buf, in)
	}
}

func serializeInput(buf *bytes.Buffer, in Input) {
	buf.Write(in.PrevTxID[:])
	writeUint32(buf, in.OutputIndex)
	writeUint32(buf, uint32(len(in.Signature)))
	writeBytes(buf, in.Signature)
	buf.Write(in.PubKey[:])
	writeUint32(buf, in.Sequence)
}

func doubleSHA256(data []byte) []byte {
	first := sha256.Sum256(data)
	second := sha256.Sum256(first[:])
	return second[:]
}
