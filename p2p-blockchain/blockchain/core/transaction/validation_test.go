package transaction

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/utxopool"
	"strconv"
	"testing"

	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/transaction"
)

type MockUTXOService struct {
	utxos map[string]transaction.Output
}

func (m *MockUTXOService) GetUTXOEntry(outpoint utxopool.Outpoint) (utxopool.UTXOEntry, error) {
	panic("implement me")
}

func (m *MockUTXOService) ContainsUTXO(outpoint utxopool.Outpoint) bool {
	panic("implement me")
}

func (m *MockUTXOService) GetUTXO(txID transaction.TransactionID, outputIndex uint32) (transaction.Output, error) {
	key := string(txID[:]) + ":" + strconv.Itoa(int(outputIndex))
	out, ok := m.utxos[key]
	if !ok {
		return transaction.Output{}, ErrUTXONotFound
	}
	return out, nil
}

func setupTestTransaction(t *testing.T) (*ecdsa.PrivateKey, transaction.Transaction, transaction.UTXO, *MockUTXOService) {
	t.Helper()

	// Generate key pair
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	// Create UTXO
	prevTxID := make([]byte, 32)
	prevIndex := uint32(0)
	pubKeyBytes := make([]byte, 33)
	compressed := elliptic.MarshalCompressed(privKey.Curve, privKey.X, privKey.Y)
	copy(pubKeyBytes, compressed)
	pubKeyHash := transaction.Hash160(transaction.PubKey(pubKeyBytes))

	utxo := transaction.UTXO{
		TxID:        transaction.TransactionID(prevTxID),
		OutputIndex: prevIndex,
		Output: transaction.Output{
			Value:      1000,
			PubKeyHash: pubKeyHash,
		},
	}

	utxos := []transaction.UTXO{utxo}

	// Create signed transaction
	tx, err := transaction.NewTransaction(utxos, pubKeyHash, 500, 10, privKey)
	if err != nil {
		t.Fatal(err)
	}

	// Create mock UTXO service
	mockUTXO := &MockUTXOService{
		utxos: map[string]transaction.Output{
			string(prevTxID) + ":" + strconv.Itoa(int(prevIndex)): utxo.Output,
		},
	}

	return privKey, *tx, utxo, mockUTXO
}

func TestValidateTransaction_Success(t *testing.T) {
	_, tx, _, mockUTXO := setupTestTransaction(t)

	validator := ValidationService{
		UTXOService: mockUTXO,
	}

	if err := validator.ValidateTransaction(&tx); err != nil {
		t.Fatal("expected transaction to validate:", err)
	}
}

func TestValidateTransaction_UTXONotFound(t *testing.T) {
	_, tx, _, _ := setupTestTransaction(t)

	brokenUTXO := &MockUTXOService{utxos: map[string]transaction.Output{}}
	brokenValidator := ValidationService{UTXOService: brokenUTXO}

	if err := brokenValidator.ValidateTransaction(&tx); !errors.Is(err, ErrUTXONotFound) {
		t.Fatalf("expected validation to fail due to missing UTXO, was %v", err)
	}
}

func TestValidateTransaction_PubKeyMismatch(t *testing.T) {
	_, tx, utxo, _ := setupTestTransaction(t)

	badUTXO := &MockUTXOService{
		utxos: map[string]transaction.Output{
			string(utxo.TxID[:]) + ":" + strconv.Itoa(int(utxo.OutputIndex)): {
				Value:      utxo.Output.Value,
				PubKeyHash: transaction.PubKeyHash{}, // wrong pubkey hash
			},
		},
	}
	badValidator := ValidationService{UTXOService: badUTXO}

	if err := badValidator.ValidateTransaction(&tx); !errors.Is(err, ErrPubKeyMismatch) {
		t.Fatal("expected validation to fail due to pubkey mismatch")
	}
}

func TestValidateTransaction_InvalidSignature(t *testing.T) {
	// Setup original UTXO and key
	_, realTransaction, utxo, mockUTXO := setupTestTransaction(t)

	// Generate a wrong private key for signing
	wrongKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	toPubKeyHash := utxo.Output.PubKeyHash
	amount := uint64(500)
	fee := uint64(10)

	fakePubKeyBytes := make([]byte, 33)
	compressed := elliptic.MarshalCompressed(wrongKey.Curve, wrongKey.X, wrongKey.Y)
	copy(fakePubKeyBytes, compressed)
	fakePubKeyHash := transaction.Hash160(transaction.PubKey(fakePubKeyBytes))

	fakeUTXO := transaction.UTXO{
		TxID:        utxo.TxID,
		OutputIndex: 0,
		Output: transaction.Output{
			Value:      1000,
			PubKeyHash: fakePubKeyHash,
		},
	}

	tx, err := transaction.NewTransaction(
		[]transaction.UTXO{fakeUTXO},
		toPubKeyHash,
		amount,
		fee,
		wrongKey,
	)

	if err != nil {
		t.Fatal(err)
	}

	realTransaction.Inputs[0].Signature = tx.Inputs[0].Signature

	// Use the original UTXO service (which expects original pubkey)
	validator := ValidationService{UTXOService: mockUTXO}

	// Validate transaction - should fail due to signature mismatch
	if err := validator.ValidateTransaction(&realTransaction); !errors.Is(err, ErrSignatureInvalid) {
		t.Fatal("expected validation to fail due to invalid signature")
	}
}

func TestValidateTransaction_FeeManipulated(t *testing.T) {
	_, tx, _, mockUTXO := setupTestTransaction(t)

	// Increase output value to break fee
	txTampered := tx
	txTampered.Outputs[0].Value += 1000

	validator := ValidationService{UTXOService: mockUTXO}
	err := validator.ValidateTransaction(&txTampered)
	if !errors.Is(err, ErrSignatureInvalid) {
		t.Fatal("expected validation to fail due to fee manipulation")
	}
}
