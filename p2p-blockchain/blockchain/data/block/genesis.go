package block

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/transaction"
	"time"
)

func GenesisBlock() Block {
	// Coinbase Transaction
	coinbaseTx := transaction.Transaction{
		Inputs: []transaction.Input{
			{
				PrevTxID:    transaction.TransactionID{}, // Zero hash
				OutputIndex: 0xFFFFFFFF,
				Signature:   []byte{0x00, 0x00}, // Arbitrary data
				Sequence:    0xFFFFFFFF,
			},
		},
		Outputs: []transaction.Output{
			{
				Value:      50 * 100000000,           // 50 coins
				PubKeyHash: transaction.PubKeyHash{}, // Burn or specific address
			},
		},
		LockTime: 0,
	}

	// Genesis Header
	header := BlockHeader{
		PrevHash:         Hash{}, // Zero hash
		MerkleRoot:       Hash{}, // Should be calculated
		Timestamp:        time.Date(2025, 12, 19, 5, 0, 0, 0, time.UTC).Unix(),
		DifficultyTarget: 0x1d00ffff, // Standard difficulty
		Nonce:            2083236893,
	}

	return Block{
		Header:       header,
		Transactions: []transaction.Transaction{coinbaseTx},
	}
}
