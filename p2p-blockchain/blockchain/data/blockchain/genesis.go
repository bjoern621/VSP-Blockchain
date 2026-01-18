package blockchain

import (
	"encoding/hex"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"slices"
	"time"

	"bjoernblessin.de/go-utils/util/assert"
)

func GenesisBlock() block.Block {
	// Coinbase Transaction
	coinbaseTx := transaction.Transaction{
		Inputs: []transaction.Input{
			{
				PrevTxID:    transaction.TransactionID{}, // Coinbase specific
				OutputIndex: 0xFFFFFFFF,                  // Coinbase specific
				Signature:   []byte("Tagesschau 19.12.2025: Worüber im Bundesrat entschieden wird - Fast 100 Tagesordnungspunkte liegen dem Bundesrat zur letzten Sitzung in diesem Jahr vor. Darunter auch die zuvor umstrittenen Projekte zu Rente, Krankenkassenfinanzen und Wehrdienst. Ein Blick auf die wichtigsten Themen."),
				PubKey:      transaction.PubKey{}, // Coinbase specific
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
	}

	// Sanity check: Verify merkle root
	merkleRoot := block.MerkleRootFromTransactions([]transaction.Transaction{coinbaseTx})
	expectedMerkleRoot, err := hex.DecodeString("6eb30762d7b80c291cab70c869fb8e7fc4f6a29295bb3f62cb82e89e9634aeb9")
	assert.IsNil(err, "failed to decode expected merkle root")
	assert.Assert(slices.Compare(merkleRoot[:], expectedMerkleRoot) == 0, "calculated merkle root does not match expected merkle root")

	// Genesis Header
	header := block.BlockHeader{
		PreviousBlockHash: common.Hash{}, // Genesis specific
		MerkleRoot:        merkleRoot,
		Timestamp:         time.Date(2025, 12, 19, 8, 0, 0, 0, time.UTC).Unix(),
		DifficultyTarget:  28, // StandardDifficultyTarget
		Nonce:             1389802128,
	}

	genesisBlock := block.Block{
		Header:       header,
		Transactions: []transaction.Transaction{coinbaseTx},
	}

	// Sanity check: Verify genesis block hash
	expectedGenesisHash, err := hex.DecodeString("00000000961cc158f2d959ef5980346066a2cc4aa7b3b0eb0d91f83fb918f292")
	assert.IsNil(err, "failed to decode expected genesis hash")
	actualGenesisHash := genesisBlock.Hash()
	assert.Assert(slices.Compare(actualGenesisHash[:], expectedGenesisHash) == 0, "genesis block hash does not match expected hash")

	return genesisBlock
}
