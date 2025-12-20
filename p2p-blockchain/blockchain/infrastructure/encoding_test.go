package infrastructure

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/transaction"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/utxopool"
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	var pubKeyHash transaction.PubKeyHash
	for i := range pubKeyHash {
		pubKeyHash[i] = byte(i * 5)
	}

	entry := utxopool.UTXOEntry{
		Output: transaction.Output{
			Value:      123456789,
			PubKeyHash: pubKeyHash,
		},
		BlockHeight: 999999,
		IsCoinbase:  true,
	}

	encoded := encodeUTXOEntry(entry)
	decoded := decodeUTXOEntry(encoded)

	if decoded.Output.Value != entry.Output.Value {
		t.Errorf("Value mismatch: got %d, want %d", decoded.Output.Value, entry.Output.Value)
	}
	if decoded.Output.PubKeyHash != entry.Output.PubKeyHash {
		t.Error("PubKeyHash mismatch")
	}
	if decoded.BlockHeight != entry.BlockHeight {
		t.Errorf("BlockHeight mismatch: got %d, want %d", decoded.BlockHeight, entry.BlockHeight)
	}
	if decoded.IsCoinbase != entry.IsCoinbase {
		t.Errorf("IsCoinbase mismatch: got %v, want %v", decoded.IsCoinbase, entry.IsCoinbase)
	}
}
