package utxo

import (
	"fmt"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/blockchain"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"

	"bjoernblessin.de/go-utils/util/assert"
	"bjoernblessin.de/go-utils/util/logger"
)

// UtxoStoreAPI provides UTXO set management for the blockchain.
type UtxoStoreAPI interface {
	// InitializeGenesisPool creates the UTXO pool for the genesis block.
	// This must be called before AddNewBlock can process any blocks.
	InitializeGenesisPool(genesisBlock block.Block) error

	// AddNewBlock updates the UTXO set with a new block's transactions.
	AddNewBlock(block block.Block) error

	// GetUtxoFromBlock retrieves a specific UTXO from a given block's UTXO pool.
	GetUtxoFromBlock(prevTxID transaction.TransactionID, outputIndex uint32, blockHash common.Hash) (transaction.Output, error)

	// ValidateTransactionsOfBlock checks if all inputs in the block's transactions reference valid UTXOs.
	// Precondition: the previous block's UTXO pool must exist. For this the previous block must have been added already.
	ValidateTransactionsOfBlock(blockToValidate block.Block) bool

	// ValidateTransactionFromBlock checks if a transaction is valid against the UTXO set at a specific block.
	ValidateTransactionFromBlock(tx transaction.Transaction, blockHash common.Hash) bool

	// GetUtxosByPubKeyHashFromBlock retrieves all UTXOs associated with a public key hash from a specific block's UTXO pool.
	GetUtxosByPubKeyHashFromBlock(pubKeyHash transaction.PubKeyHash, blockHash common.Hash) ([]transaction.UTXO, error)
}

// outpoint uniquely identifies a UTXO by transaction ID and output index
type outpoint struct {
	TxID        transaction.TransactionID
	OutputIndex uint32
}

// utxoPool holds the UTXO set for a specific block.
type utxoPool struct {
	UtxoData map[outpoint]transaction.Output
}

// UtxoStore manages UTXO pools for each block in the blockchain.
type UtxoStore struct {
	blockHashToPool map[common.Hash]utxoPool // maps block hash to its UTXO pool
	blockStore      blockchain.BlockStoreAPI
}

// NewUtxoStore creates a new UTXO store with the given block store.
func NewUtxoStore(blockStore blockchain.BlockStoreAPI) UtxoStoreAPI {
	return &UtxoStore{
		blockHashToPool: make(map[common.Hash]utxoPool),
		blockStore:      blockStore,
	}
}

// InitializeGenesisPool creates the UTXO pool for the genesis block.
// This must be called before AddNewBlock can process any blocks.
// The genesis block's previous block hash is empty, so we create an empty previous pool
// and then build the genesis pool from it.
func (us *UtxoStore) InitializeGenesisPool(genesisBlock block.Block) error {
	genesisHash := genesisBlock.Hash()

	// Check if already initialized
	if _, exists := us.blockHashToPool[genesisHash]; exists {
		logger.Warnf("[utxoStore] genesis block %v already exists in UTXO store, skipping", genesisHash)
		return nil
	}

	// Create an empty pool for the "previous" block (which doesn't exist for genesis)
	emptyPrevPool := &utxoPool{UtxoData: make(map[outpoint]transaction.Output)}

	// Build the genesis pool using the standard method
	genesisPool := us.createUtxoPoolFromBlock(emptyPrevPool, genesisBlock)

	// Store the genesis pool
	us.blockHashToPool[genesisHash] = *genesisPool

	logger.Infof("[utxoStore] initialized genesis block UTXO pool with %d UTXOs", len(genesisPool.UtxoData))

	return nil
}

// ValidateTransactionFromBlock checks if all inputs in the transaction reference valid UTXOs
// from the UTXO pool at the specified block hash.
func (us *UtxoStore) ValidateTransactionFromBlock(tx transaction.Transaction, blockHash common.Hash) bool {
	prevPool, exists := us.blockHashToPool[blockHash]
	if !exists {
		return false // Previous block's UTXO pool not found; block is invalid
	}

	for _, input := range tx.Inputs {
		outpoint := outpoint{
			TxID:        input.PrevTxID,
			OutputIndex: input.OutputIndex,
		}
		if _, exists := prevPool.UtxoData[outpoint]; !exists {
			return false // Referenced UTXO not found; block is invalid
		}
	}

	return true
}

// ValidateTransactionsOfBlock checks if all inputs in the block's transactions reference valid UTXOs.
// Precondition: the previous block's UTXO pool must exist. For this the previous block must have been added already.
func (us *UtxoStore) ValidateTransactionsOfBlock(blockToValidate block.Block) bool {
	for _, tx := range blockToValidate.Transactions {
		valid := us.ValidateTransactionFromBlock(tx, blockToValidate.Header.PreviousBlockHash)
		if !valid {
			return false // Invalid transaction found
		}
	}

	return true // All inputs are valid
}

// GetUtxoFromBlock retrieves a specific UTXO from a given block's UTXO pool.
func (us *UtxoStore) GetUtxoFromBlock(id transaction.TransactionID, outputIndex uint32, blockHash common.Hash) (transaction.Output, error) {
	blockPool, exists := us.blockHashToPool[blockHash]
	if !exists {
		return transaction.Output{}, fmt.Errorf("UTXO pool for block %v not found", blockHash)
	}

	outpoint := outpoint{
		TxID:        id,
		OutputIndex: outputIndex,
	}

	output, exists := blockPool.UtxoData[outpoint]
	if !exists {
		return transaction.Output{}, fmt.Errorf("UTXO %v:%d not found in block %v", id, outputIndex, blockHash)
	}

	return output, nil
}

// GetUtxosByPubKeyHashFromBlock retrieves all UTXOs associated with a public key hash from a specific block's UTXO pool.
func (us *UtxoStore) GetUtxosByPubKeyHashFromBlock(pubKeyHash transaction.PubKeyHash, blockHash common.Hash) ([]transaction.UTXO, error) {
	blockPool, exists := us.blockHashToPool[blockHash]
	if !exists {
		return nil, fmt.Errorf("UTXO pool for block %v not found", blockHash)
	}

	utxos := make([]transaction.UTXO, 0)

	for outpoint, output := range blockPool.UtxoData {
		if output.PubKeyHash == pubKeyHash {
			utxo := transaction.UTXO{
				TxID:        outpoint.TxID,
				OutputIndex: outpoint.OutputIndex,
				Output:      output,
			}
			utxos = append(utxos, utxo)
		}
	}

	return utxos, nil
}

// AddNewBlock creates a new UTXO pool for the block by removing spent UTXOs and adding new outputs.
// Skips orphan blocks and blocks already in the store.
func (us *UtxoStore) AddNewBlock(newBlock block.Block) error {
	logger.Infof("[utxoStore] adding new block %v to UTXO store", newBlock.Header.Hash())

	newBlockHash := newBlock.Header.Hash()
	if isOrphan, err := us.blockStore.IsOrphanBlock(newBlock); isOrphan {
		logger.Warnf("[utxoStore] block %v is an orphan block: %v, skipping", newBlockHash, err)
		return nil
	}

	if _, exists := us.blockHashToPool[newBlockHash]; exists {
		logger.Warnf("[utxoStore] block %v already exists in UTXO store, skipping", newBlockHash)
		return nil
	}

	prevBlockHash := newBlock.Header.PreviousBlockHash
	prevPool, exists := us.blockHashToPool[prevBlockHash]
	assert.Assert(exists, "previous block UTXO pool not found, but must exist, as block is no orphan")

	valid := us.ValidateTransactionsOfBlock(newBlock)
	if !valid {
		return fmt.Errorf("block %v is invalid, cannot add to UTXO store", newBlockHash)
	}

	newPool := us.createUtxoPoolFromBlock(&prevPool, newBlock)

	// Store the new pool
	us.blockHashToPool[newBlockHash] = *newPool

	return nil
}

// createUtxoPoolFromBlock builds a new UTXO pool by copying the previous pool,
// removing UTXOs spent by the block's inputs, and adding the block's new outputs.
func (us *UtxoStore) createUtxoPoolFromBlock(prevPool *utxoPool, newBlock block.Block) *utxoPool {
	newUtxoData := make(map[outpoint]transaction.Output)

	// Add all previous UTXOs to the new pool
	for outpoint, output := range prevPool.UtxoData {
		newUtxoData[outpoint] = output
	}

	// Remove all spent UTXOs from the new block
	for _, tx := range newBlock.Transactions {
		// Skip coinbase transaction inputs (they don't reference real UTXOs)
		if tx.IsCoinbase() {
			continue
		}

		for _, input := range tx.Inputs {
			outpoint := outpoint{
				TxID:        input.PrevTxID,
				OutputIndex: input.OutputIndex,
			}
			_, exists := newUtxoData[outpoint]
			assert.Assert(exists, "referenced UTXO not found in pool")
			delete(newUtxoData, outpoint)
		}
	}

	// Add all new outputs from the new block
	for _, tx := range newBlock.Transactions {
		txID := tx.TransactionId()
		for i, output := range tx.Outputs {
			outpoint := outpoint{
				TxID:        txID,
				OutputIndex: uint32(i),
			}
			newUtxoData[outpoint] = output
		}
	}

	return &utxoPool{
		UtxoData: newUtxoData,
	}
}
