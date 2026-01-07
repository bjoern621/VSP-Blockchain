package miner

import (
	"encoding/hex"
	"fmt"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"testing"
	"time"

	"bjoernblessin.de/go-utils/util/assert"
)

func TestBlockchain_Inv_InvokesRequestDataByCallingSendGetData(t *testing.T) {
	gen := GenesisBlock()

	m := MinerService{}
	nonce := m.MineBlock(gen)
	fmt.Printf("nonce: %d", nonce)
}

func GenesisBlock() block.Block {
	// Coinbase Transaction
	coinbaseTx := transaction.Transaction{
		Inputs: []transaction.Input{
			{
				PrevTxID:    transaction.TransactionID{}, // Coinbase specific
				OutputIndex: 0xFFFFFFFF,                  // Coinbase specific
				Signature:   []byte("Tagesschau 19.12.2025: Worüber im Bundesrat entschieden wird - Fast 100 Tagesordnungspunkte liegen dem Bundesrat zur letzten Sitzung in diesem Jahr vor. Darunter auch die zuvor umstrittenen Projekte zu Rente, Krankenkassenfinanzen und Wehrdienst. Ein Blick auf die wichtigsten Themen."),
				PubKey:      transaction.PubKey{}, // Coinbase specific
				Sequence:    0xFFFFFFFF,
			},
		},
		Outputs: []transaction.Output{
			{
				Value: 50,
				// VS Address: "18pDpp7xK8VBx8egZmWCpnbVnEZcPRT9LA" (Base58Check)
				// Calculated using keys.NewKeyEncodingsImpl().Base58CheckToBytes()
				PubKeyHash: [common.PublicKeyHashSize]byte{0x55, 0xb7, 0x22, 0xcb, 0xae, 0x36, 0xb5, 0x8c, 0x2a, 0xd2, 0xf2, 0xdd, 0x64, 0x06, 0x35, 0xdc, 0x05, 0x42, 0x45, 0xe4},
			},
		},
		LockTime: 0,
	}

	// Verify merkle root

	expectedMerkleRoot, err := hex.DecodeString("0cff1e741712b5b4671a88918dd003aaed08a76e234bd484b42e169dd844db03")
	assert.IsNil(err, "failed to decode expected merkle root")

	//assert.Assert(slices.Compare(merkleRoot[:], expectedMerkleRoot) == 0, "calculated merkle root does not match expected merkle root")

	var hash common.Hash
	copy(hash[:], expectedMerkleRoot[:])

	// Genesis Header
	header := block.BlockHeader{
		PreviousBlockHash: common.Hash{}, // Genesis specific
		MerkleRoot:        hash,
		Timestamp:         time.Date(2025, 12, 19, 8, 0, 0, 0, time.UTC).Unix(),
		DifficultyTarget:  24,
		Nonce:             0,
	}

	return block.Block{
		Header:       header,
		Transactions: []transaction.Transaction{coinbaseTx},
	}
}
