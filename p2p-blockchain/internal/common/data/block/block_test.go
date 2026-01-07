package block

import (
	"bytes"
	"reflect"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
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

func TestMerkleRoot_OneTx(t *testing.T) {
	tx := transaction.Transaction{} // zero value should still hash deterministically
	b := Block{Transactions: []transaction.Transaction{tx}}

	got := b.MerkleRoot()

	h := tx.Hash()
	want := doubleSHA256(append(h[:], h[:]...)) // duplicated leaf when odd count

	if got != want {
		t.Fatalf("merkle root mismatch\n got:  %x\n want: %x", got, want)
	}
}

func TestMerkleRoot_TwoTx(t *testing.T) {
	tx1 := transaction.Transaction{}
	tx2 := transaction.Transaction{}
	b := Block{Transactions: []transaction.Transaction{tx1, tx2}}

	got := b.MerkleRoot()

	h1 := tx1.Hash()
	h2 := tx2.Hash()
	want := doubleSHA256(append(h1[:], h2[:]...))

	if got != want {
		t.Fatalf("merkle root mismatch\n got:  %x\n want: %x", got, want)
	}
}

func TestMerkleRoot_ThreeTx_DuplicatesLast(t *testing.T) {
	tx1 := transaction.Transaction{}
	tx2 := transaction.Transaction{}
	tx3 := transaction.Transaction{}
	b := Block{Transactions: []transaction.Transaction{tx1, tx2, tx3}}

	got := b.MerkleRoot()

	// Leaves: h1, h2, h3, h3
	h1 := tx1.Hash()
	h2 := tx2.Hash()
	h3 := tx3.Hash()

	p12 := doubleSHA256(append(h1[:], h2[:]...))
	p33 := doubleSHA256(append(h3[:], h3[:]...))
	want := doubleSHA256(append(p12[:], p33[:]...))

	if got != want {
		t.Fatalf("merkle root mismatch\n got:  %x\n want: %x", got, want)
	}
}

// More complex uneven test: 9 transactions cause duplication at multiple tree levels:
// 9 -> 10 -> 5 -> 6 -> 3 -> 4 -> 2 -> 1
func TestMerkleRoot_NineTx_UnevenMultipleLevels(t *testing.T) {
	txs := make([]transaction.Transaction, 0, 9)
	for i := 0; i < 9; i++ {
		txs = append(txs, makeTxWithTag(byte(i+1)))
	}

	b := Block{Transactions: txs}
	got := b.MerkleRoot()

	leaves := make([]common.Hash, 0, len(txs))
	for _, tx := range txs {
		leaves = append(leaves, tx.Hash())
	}
	want := merkleRootReferenceFromLeaves(leaves)

	if got != want {
		t.Fatalf("merkle root mismatch\n got:  %x\n want: %x", got, want)
	}
}

// makeTxWithTag tries to make transactions differ without relying on concrete struct fields.
// It sets the first settable exported field it can find (string/int/uint/[]byte) based on tag.
func makeTxWithTag(tag byte) transaction.Transaction {
	tx := transaction.Transaction{}
	v := reflect.ValueOf(&tx).Elem()
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		if !f.CanSet() {
			continue
		}

		switch f.Kind() {
		case reflect.String:
			f.SetString(string([]byte{tag}))
			return tx
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			f.SetInt(int64(tag))
			return tx
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			f.SetUint(uint64(tag))
			return tx
		case reflect.Slice:
			// try []byte
			if f.Type().Elem().Kind() == reflect.Uint8 {
				f.SetBytes([]byte{tag, tag ^ 0xFF, tag + 1})
				return tx
			}
		case reflect.Array:
			// try [N]byte
			if f.Type().Elem().Kind() == reflect.Uint8 {
				n := f.Len()
				for j := 0; j < n; j++ {
					f.Index(j).SetUint(uint64(tag + byte(j)))
				}
				return tx
			}
		}
	}
	// If we can't set anything, we still return the zero tx; the test will still validate the algorithm.
	return tx
}

func merkleRootReferenceFromLeaves(leaves []common.Hash) common.Hash {
	if len(leaves) == 0 {
		return common.Hash{}
	}

	hashes := make([]common.Hash, len(leaves))
	copy(hashes, leaves)

	if len(hashes)%2 == 1 {
		hashes = append(hashes, hashes[len(hashes)-1])
	}

	for len(hashes) != 1 {
		if len(hashes)%2 == 1 {
			hashes = append(hashes, hashes[len(hashes)-1])
		}

		next := make([]common.Hash, 0, len(hashes)/2)
		for i := 0; i < len(hashes); i += 2 {
			combined := append(hashes[i][:], hashes[i+1][:]...)
			next = append(next, doubleSHA256(combined))
		}
		hashes = next
	}

	return hashes[0]
}

func makeTx(lockTime uint64) transaction.Transaction {
	return transaction.Transaction{
		Inputs:   nil,
		Outputs:  nil,
		LockTime: lockTime,
	}
}
