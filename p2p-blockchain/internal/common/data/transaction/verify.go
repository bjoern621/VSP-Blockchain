package transaction

import (
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
)

// VerifySignature verifies the signature of a specific input in the transaction.
// It returns true if the signature is valid, false otherwise.
// The referencedOutput is the UTXO being spent by this input.
func (tx *Transaction) VerifySignature(inputIndex int, referencedOutput Output) (bool, error) {
	if inputIndex < 0 || inputIndex >= len(tx.Inputs) {
		return false, nil
	}

	input := tx.Inputs[inputIndex]

	// Compute the signature hash for this input
	sighash, err := tx.SigHash(inputIndex, referencedOutput)
	if err != nil {
		return false, err
	}

	// Parse the public key from the input
	pubKey, err := btcec.ParsePubKey(input.PubKey[:])
	if err != nil {
		return false, err
	}

	// Parse the DER-encoded signature
	sig, err := ecdsa.ParseDERSignature(input.Signature)
	if err != nil {
		return false, err
	}

	// Verify the signature
	return sig.Verify(sighash, pubKey), nil
}

// VerifyAllSignatures verifies all input signatures in the transaction.
// It requires the referenced outputs (UTXOs) for each input to compute the signature hash.
// Returns true if all signatures are valid, false otherwise.
func (tx *Transaction) VerifyAllSignatures(referencedOutputs []Output) (bool, error) {
	if len(referencedOutputs) != len(tx.Inputs) {
		return false, nil
	}

	for i := range tx.Inputs {
		valid, err := tx.VerifySignature(i, referencedOutputs[i])
		if err != nil {
			return false, err
		}
		if !valid {
			return false, nil
		}
	}

	return true, nil
}
