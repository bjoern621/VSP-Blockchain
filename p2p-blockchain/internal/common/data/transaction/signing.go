package transaction

import (
	"errors"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
)

// Sign calculates the Signature for all Inputs in the Transaction
func (tx *Transaction) Sign(privateKey PrivateKey, utxos []UTXO) error {
	for i := range tx.Inputs {
		err := tx.signInput(privateKey, utxos, i)
		if err != nil {
			return err
		}
	}
	return nil
}

func (tx *Transaction) signInput(privateKey PrivateKey, utxos []UTXO, inputIndex int) error {

	input := &tx.Inputs[inputIndex]

	referenced, ok := findUTXO(input.PrevTxID, input.OutputIndex, utxos)
	if !ok {
		return errors.New("UTXO not found")
	}

	sighash, err := tx.SigHash(inputIndex, referenced)
	if err != nil {
		return err
	}

	privKey, pubKey := btcec.PrivKeyFromBytes(privateKey[:])

	sig := ecdsa.Sign(privKey, sighash)

	// DER encode
	derSig := sig.Serialize()
	input.Signature = derSig

	copy(input.PubKey[:], pubKey.SerializeCompressed())

	return nil
}

func findUTXO(id TransactionID, index uint32, utxos []UTXO) (Output, bool) {
	for _, u := range utxos {
		if u.TxID == id && u.OutputIndex == index {
			return u.Output, true
		}
	}
	return Output{}, false
}
