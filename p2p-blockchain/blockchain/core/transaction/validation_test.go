package transaction

import (
	"errors"
	"strconv"
	"testing"

	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/transaction"

	"github.com/btcsuite/btcd/btcec/v2"
)

type MockUTXOService struct {
	utxos map[string]transaction.Output
}

func (m *MockUTXOService) GetUTXO(txID transaction.TransactionID, index uint32) (transaction.Output, bool) {
	key := string(txID[:]) + ":" + strconv.Itoa(int(index))
	out, ok := m.utxos[key]
	return out, ok
}

func setupTestTransaction(
	t *testing.T,
) (transaction.PrivateKey, transaction.Transaction, transaction.UTXO, *MockUTXOService) {

	t.Helper()

	// Generate secp256k1 key
	btcecPriv, err := btcec.NewPrivateKey()
	if err != nil {
		t.Fatal(err)
	}

	var privKey transaction.PrivateKey
	copy(privKey[:], btcecPriv.Serialize())

	// Create UTXO
	prevTxID := make([]byte, 32)
	prevIndex := uint32(0)

	pubKeyBytes := btcecPriv.PubKey().SerializeCompressed()
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
	tx, err2 := transaction.NewTransaction(
		utxos,
		pubKeyHash,
		500,
		10,
		privKey,
	)
	if err2 != nil {
		t.Fatal(err2)
	}

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
		t.Fatal("expected validation to fail due to missing UTXO")
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

	wrongBtcecKey, err := btcec.NewPrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	var wrongKey transaction.PrivateKey
	copy(wrongKey[:], wrongBtcecKey.Serialize())

	toPubKeyHash := utxo.Output.PubKeyHash
	amount := uint64(500)
	fee := uint64(10)

	fakePubKeyBytes := wrongBtcecKey.PubKey().SerializeCompressed()
	fakePubKeyHash := transaction.Hash160(transaction.PubKey(fakePubKeyBytes))

	fakeUTXO := transaction.UTXO{
		TxID:        utxo.TxID,
		OutputIndex: 0,
		Output: transaction.Output{
			Value:      1000,
			PubKeyHash: fakePubKeyHash,
		},
	}

	tx, err2 := transaction.NewTransaction(
		[]transaction.UTXO{fakeUTXO},
		toPubKeyHash,
		amount,
		fee,
		wrongKey,
	)
	if err2 != nil {
		t.Fatal(err2)
	}

	// Inject invalid signature
	realTransaction.Inputs[0].Signature = tx.Inputs[0].Signature

	validator := ValidationService{UTXOService: mockUTXO}

	if err = validator.ValidateTransaction(&realTransaction); !errors.Is(err, ErrSignatureInvalid) {
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
