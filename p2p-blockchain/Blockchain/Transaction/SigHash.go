package Transaction

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
)

func (tx *Transaction) SigHash(inputIndex int, referenced Output) ([]byte, error) {
	if inputIndex >= len(tx.Inputs) {
		return nil, errors.New("input index out of range")
	}

	buf := new(bytes.Buffer)

	addInputList(buf, inputIndex, referenced, tx)

	addOutputList(buf, tx)

	writeUint64(buf, tx.LockTime)

	return doubleSHA256Hash(buf)
}

func doubleSHA256Hash(buf *bytes.Buffer) ([]byte, error) {
	first := sha256.Sum256(buf.Bytes())
	second := sha256.Sum256(first[:])
	return second[:], nil
}

func addOutputList(buf *bytes.Buffer, tx *Transaction) {
	writeUint32(buf, uint32(len(tx.Outputs)))
	for _, out := range tx.Outputs {
		addOutput(buf, out)
	}
}

func addOutput(buf *bytes.Buffer, out Output) {
	writeUint64(buf, out.Value)
	writeBytes(buf, out.PubKeyHash[:])
}

func addInputList(buf *bytes.Buffer, inputIndex int, referenced Output, tx *Transaction) {
	writeUint32(buf, uint32(len(tx.Inputs)))
	for i, in := range tx.Inputs {
		var inputToBeSigned = inputIndex == i
		addInput(buf, referenced, in, inputToBeSigned)
	}
}

func addInput(buf *bytes.Buffer, referenced Output, in Input, toBeSigned bool) {
	buf.Write(in.PrevTxID[:])
	writeUint32(buf, in.OutputIndex)
	if toBeSigned {
		writeBytes(buf, referenced.PubKeyHash[:])
		writeUint64(buf, referenced.Value)
	} else {
		writeUint64(buf, uint64(0))
	}

	writeUint32(buf, in.Sequence)
}

/*
* Helpers are being used to ignore Error for Write() since bytes.Buffer doesnt fail
 */
func writeBytes(buf *bytes.Buffer, data []byte) {
	_, _ = buf.Write(data)
}

func writeUint32(buf *bytes.Buffer, val uint32) {
	_ = binary.Write(buf, binary.LittleEndian, val)
}

func writeUint64(buf *bytes.Buffer, val uint64) {
	_ = binary.Write(buf, binary.LittleEndian, val)
}
