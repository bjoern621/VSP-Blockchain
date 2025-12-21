package block

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"

	"bjoernblessin.de/go-utils/util/assert"
)

// Block represents a block in the blockchain.
// It consists of a BlockHeader and a list of Transactions.
type Block struct {
	Header       BlockHeader
	Transactions []transaction.Transaction
}

func (b *Block) Hash() common.Hash {
	merkleRoot := b.MerkleRoot()

	assert.Assert(bytes.Equal(merkleRoot[:], b.Header.MerkleRoot[:]), "Merkle root of block does not match header")

	var buffer = make([]byte, 0)
	buffer = append(buffer, merkleRoot[:]...)
	buffer = append(buffer, b.Header.PreviousBlockHash[:]...)
	buffer = binary.LittleEndian.AppendUint64(buffer, uint64(b.Header.Timestamp))
	buffer = binary.LittleEndian.AppendUint32(buffer, b.Header.DifficultyTarget)
	buffer = binary.LittleEndian.AppendUint32(buffer, b.Header.Nonce)

	return doubleSHA256(buffer)
}

func (b *Block) MerkleRoot() common.Hash {
	tmpTransactions := make([]transaction.Transaction, len(b.Transactions))
	copy(tmpTransactions, b.Transactions)

	if len(b.Transactions)%2 == 1 {
		// Append last transaction to make event number of transactions
		tmpTransactions = append(tmpTransactions, b.Transactions[len(b.Transactions)-1])
	}

	var hashes = make([]common.Hash, len(tmpTransactions))
	for i, tx := range tmpTransactions {
		hashes[i] = tx.Hash()
	}

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
