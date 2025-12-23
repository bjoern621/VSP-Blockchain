package validation

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/asn1"
	"errors"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/utxo"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
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
	UTXOService utxo.LookupAPI
}

type ValidationAPI interface {
	ValidateTransaction(tx *transaction.Transaction) (bool, error)
}

// ValidateTransaction validates all inputs in a transaction by checking if each of the given inputs exists and all signatures are valid.
func (v *ValidationService) ValidateTransaction(tx *transaction.Transaction) (bool, error) {
	for i := range tx.Inputs {
		if err := v.validateInput(tx, i); err != nil {
			return false, err
		}
	}
	return true, nil
}

// validateInput validates a single input
func (v *ValidationService) validateInput(tx *transaction.Transaction, inputIndex int) error {
	in := tx.Inputs[inputIndex]

	referenced, err := v.getReferencedUTXO(in)
	if err != nil {
		return err
	}

	// Compute sighash
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
	// Decode DER signature
	var sig transaction.EcdsaSignature
	_, err := asn1.Unmarshal(in.Signature, &sig)
	if err != nil {
		return ErrInvalidSigEncoding
	}

	// Recover public key from compressed format
	x, y := elliptic.UnmarshalCompressed(elliptic.P256(), in.PubKey[:])
	if x == nil || y == nil {
		return ErrInvalidPubKey
	}
	pubKey := &ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}

	// Verify signature
	if !ecdsa.Verify(pubKey, sighash, sig.R, sig.S) {
		return ErrSignatureInvalid
	}
	return nil
}

func (v *ValidationService) getReferencedUTXO(in transaction.Input) (transaction.Output, error) {
	referenced, err := v.UTXOService.GetUTXO(in.PrevTxID, in.OutputIndex)
	if err != nil {
		return transaction.Output{}, ErrUTXONotFound
	}
	if transaction.Hash160(in.PubKey) != referenced.PubKeyHash {
		return transaction.Output{}, ErrPubKeyMismatch
	}
	return referenced, nil
}
