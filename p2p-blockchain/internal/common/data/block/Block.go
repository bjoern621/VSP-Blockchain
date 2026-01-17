package block

import (
	"crypto/sha256"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"

	"bjoernblessin.de/go-utils/util/assert"
	"bjoernblessin.de/go-utils/util/logger"
)

// Block represents a block in the blockchain.
// It consists of a BlockHeader and a list of Transactions.
type Block struct {
	Header BlockHeader
	// Transactions holds all transactions included in the block.
	// The first transaction is always the coinbase transaction.
	// This slice should never be empty.
	Transactions []transaction.Transaction
}

// BlockDifficulty calculates and returns the difficulty of the block.
//
// The difficulty is the number of leading zero bits in the block hash.
// This can be different from the difficulty target specified in the block header.
// Invariant: BlockDifficulty >= Header.DifficultyTarget.
//
// (In Bitcoin a different formula is used to calculate the difficulty.)
func (b *Block) BlockDifficulty() uint8 {
	hash := b.Hash()
	return countLeadingZeroBits(hash)
}

func countLeadingZeroBits(hash common.Hash) uint8 {
	var difficulty uint16 = 0

	// For each byte in the hash
	for _, byteVal := range hash {

		// Go through each bit in the byte, from most significant to least significant
		for bitPos := 7; bitPos >= 0; bitPos-- {
			bitmask := byte(1 << bitPos)

			if (byteVal & bitmask) == 0 {
				difficulty++
			} else {
				assert.Assert(difficulty <= 255, "difficulty should never exceed 255")
				return uint8(difficulty)
			}
		}
	}

	// All bits are zero, so difficulty is 256, but we cap it at 255
	// This case is considered as "255 leading zeros and one bit that is either set or not set"
	if difficulty > 255 {
		difficulty = 255
	}

	return uint8(difficulty)
}

func (b *Block) Hash() common.Hash {
	return b.Header.Hash()
}

// MerkleRoot calculates and returns the Merkle root of the block's transactions.
func (b *Block) MerkleRoot() common.Hash {
	return MerkleRootFromTransactions(b.Transactions)
}

// MerkleRootFromTransactions calculates the Merkle root from a list of transactions.
func MerkleRootFromTransactions(txs []transaction.Transaction) common.Hash {
	logger.Infof("Calculating merkle root for %d transactions", len(txs))
	tmpTransactions := make([]transaction.Transaction, len(txs))
	copy(tmpTransactions, txs)

	if len(txs)%2 == 1 {
		// Append last transaction to make even number of transactions
		tmpTransactions = append(tmpTransactions, txs[len(txs)-1])
	}

	var hashes = make([]common.Hash, len(tmpTransactions))
	for i, tx := range tmpTransactions {
		hashes[i] = tx.Hash()
	}

	return merkleRootFromHashes(hashes)
}

// merkleRootFromHashes calculates the Merkle root from a list of hashes.
// The list of hashes must have an even length.
// There must be at least one hash (the coinbase transaction).
func merkleRootFromHashes(hashes []common.Hash) common.Hash {
	assert.Assert(len(hashes) >= 1, "merkleRootFromHashes requires at least one hash (at least coinbase transaction)")
	assert.Assert(len(hashes)%2 == 0, "merkleRootFromHashes requires even number of hashes")

	for len(hashes) != 1 {
		var tmpHashes = make([]common.Hash, 0)

		// Append last hash to make event number of hashes again
		if len(hashes)%2 == 1 {
			hashes = append(hashes, hashes[len(hashes)-1])
		}

		for i := 0; i < len(hashes); i += 2 {
			combinedData := append(hashes[i][:], hashes[i+1][:]...)
			tmpHashes = append(tmpHashes, doubleSHA256(combinedData))
		}

		hashes = tmpHashes
	}

	return hashes[0]
}

func doubleSHA256(data []byte) common.Hash {
	first := sha256.Sum256(data)
	second := sha256.Sum256(first[:])
	var hash common.Hash
	copy(hash[:], second[:])
	return hash
}

// BlockWithMetadata contains a block and its metadata for visualization purposes.
type BlockWithMetadata struct {
	Block           Block
	Height          uint64
	AccumulatedWork uint64
	ParentHash      *common.Hash // nil if no parent (genesis or orphan root)
	IsOrphan        bool
	IsMainChain     bool
}
