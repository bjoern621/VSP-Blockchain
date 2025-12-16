package transaction

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/asn1"
	"errors"
	"math/big"
)

// Sign calculates the Signature for all Inputs in the Transaction
func (tx *Transaction) Sign(privateKey *ecdsa.PrivateKey, utxos []UTXO) error {
	for i := range tx.Inputs {
		err := tx.signInput(privateKey, utxos, i)
		if err != nil {
			return err
		}
	}
	return nil
}

func (tx *Transaction) signInput(privateKey *ecdsa.PrivateKey, utxos []UTXO, inputIndex int) error {
	input := &tx.Inputs[inputIndex]

	referenced, ok := findUTXO(input.PrevTxID, input.OutputIndex, utxos)
	if !ok {
		return errors.New("UTXO not found")
	}

	sighash, err := tx.SigHash(inputIndex, referenced)
	if err != nil {
		return err
	}

	r, s, err := ecdsa.Sign(rand.Reader, privateKey, sighash)
	if err != nil {
		return err
	}

	// DER-encode signature
	sig, err := asn1.Marshal(EcdsaSignature{R: r, S: s})
	if err != nil {
		return err
	}
	input.Signature = sig

	compressedPublicKeyBytes := elliptic.MarshalCompressed(privateKey.Curve, privateKey.X, privateKey.Y)
	copy(input.PubKey[:], compressedPublicKeyBytes)
	return nil
}

type EcdsaSignature struct {
	R, S *big.Int
}

func findUTXO(id TransactionID, index uint32, utxos []UTXO) (Output, bool) {
	for _, u := range utxos {
		if u.TxID == id && u.OutputIndex == index {
			return u.Output, true
		}
	}
	return Output{}, false
}
