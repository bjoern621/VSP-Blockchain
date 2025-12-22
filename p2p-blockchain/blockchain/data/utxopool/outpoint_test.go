package utxopool

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"testing"
)

func TestOutpointKey(t *testing.T) {
	var txID transaction.TransactionID
	for i := range txID {
		txID[i] = byte(i)
	}

	outpoint := NewOutpoint(txID, 42)
	key := outpoint.Key()

	reconstructed := OutpointFromKey(key)

	if reconstructed.TxID != txID {
		t.Errorf("TxID mismatch: got %v, want %v", reconstructed.TxID, txID)
	}
	if reconstructed.OutputIndex != 42 {
		t.Errorf("OutputIndex mismatch: got %d, want 42", reconstructed.OutputIndex)
	}
}
