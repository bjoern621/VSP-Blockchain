package transaction

import (
	"errors"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/transaction"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
)

var (
	ErrUTXONotFound       = errors.New("UTXO not found")
	ErrPubKeyMismatch     = errors.New("pubkey does not match output hash")
	ErrSignatureInvalid   = errors.New("signature verification failed")
	ErrInvalidPubKey      = errors.New("invalid compressed public key")
	ErrInvalidSigEncoding = errors.New("invalid signature encoding")
)

// ValidationService validates transactions using a UTXO lookup service
type ValidationService struct {
	UTXOService core.UTXOLookupService
}

// ValidateTransaction validates all inputs in a transaction by checking if each of the given inputs exists and all signatures are valid.
func (v *ValidationService) ValidateTransaction(tx *transaction.Transaction) error {
	for i := range tx.Inputs {
		if err := v.validateInput(tx, i); err != nil {
			return err
		}
	}
	return nil
}

// validateInput validates a single input
func (v *ValidationService) validateInput(tx *transaction.Transaction, inputIndex int) error {
	in := tx.Inputs[inputIndex]

	referenced, err := v.getReferencedUTXO(in)
	if err != nil {
		return err
	}

	sighash, err := tx.SigHash(inputIndex, referenced)
	if err != nil {
		return err
	}

	err2 := v.verifySignature(in, sighash)
	if err2 != nil {
		return err2
	}

	return nil
}

func (v *ValidationService) verifySignature(in transaction.Input, sighash []byte) error {
	sig, err := ecdsa.ParseDERSignature(in.Signature)
	if err != nil {
		return ErrInvalidSigEncoding
	}

	pubKey, err2 := btcec.ParsePubKey(in.PubKey[:])
	if err2 != nil {
		return ErrInvalidPubKey
	}

	if !sig.Verify(sighash, pubKey) {
		return ErrSignatureInvalid
	}

	return nil
}

func (v *ValidationService) getReferencedUTXO(in transaction.Input) (transaction.Output, error) {
	// Lookup UTXO
	referenced, ok := v.UTXOService.GetUTXO(in.PrevTxID, in.OutputIndex)
	if !ok {
		return transaction.Output{}, ErrUTXONotFound
	}
	if transaction.Hash160(in.PubKey) != referenced.PubKeyHash {
		return transaction.Output{}, ErrPubKeyMismatch
	}
	return referenced, nil
}
