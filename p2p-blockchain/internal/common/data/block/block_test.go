package block

import (
	"bytes"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"testing"
)

func TestEvenBlock(t *testing.T) {
	txs := []transaction.Transaction{
		makeTx(1),
		makeTx(2),
		makeTx(3),
		makeTx(4),
	}

	block := Block{
		Transactions: txs,
	}
	calculatedMerkleRoot := block.MerkleRoot()

	//Do calculation by hand
	h1 := txs[0].Hash()
	h2 := txs[1].Hash()
	d1_2 := append(h1[:], h2[:]...)

	h3 := txs[2].Hash()
	h4 := txs[3].Hash()
	d3_4 := append(h3[:], h4[:]...)

	r1 := doubleSHA256(d1_2)
	r2 := doubleSHA256(d3_4)

	data := append(r1[:], r2[:]...)

	expectedMerkleRoot := doubleSHA256(data)

	if !bytes.Equal(calculatedMerkleRoot[:], expectedMerkleRoot[:]) {
		t.Fatalf("Merkle root calculation failed")
	}
}

func TestUnevenBlock(t *testing.T) {
	txs := []transaction.Transaction{
		makeTx(1),
		makeTx(2),
		makeTx(3),
		makeTx(4),
		makeTx(5),
	}

	block := Block{
		Transactions: txs,
	}
	calculatedMerkleRoot := block.MerkleRoot()

	//Do calculation by hand
	h1 := txs[0].Hash()
	h2 := txs[1].Hash()
	d1_2 := append(h1[:], h2[:]...)

	h3 := txs[2].Hash()
	h4 := txs[3].Hash()
	d3_4 := append(h3[:], h4[:]...)

	h5 := txs[4].Hash()
	h6 := txs[4].Hash()
	d5_6 := append(h5[:], h6[:]...)

	r1 := doubleSHA256(d1_2)
	r2 := doubleSHA256(d3_4)
	r3 := doubleSHA256(d5_6)

	d1 := append(r1[:], r2[:]...)
	d2 := append(r3[:], r3[:]...)

	r4 := doubleSHA256(d1)
	r5 := doubleSHA256(d2)

	data := append(r4[:], r5[:]...)

	expectedMerkleRoot := doubleSHA256(data)

	if !bytes.Equal(calculatedMerkleRoot[:], expectedMerkleRoot[:]) {
		t.Fatalf("Merkle root calculation failed")
	}
}

func makeTx(lockTime uint64) transaction.Transaction {
	return transaction.Transaction{
		Inputs:   nil,
		Outputs:  nil,
		LockTime: lockTime,
	}
}
