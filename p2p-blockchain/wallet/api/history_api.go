package api

import (
	"encoding/hex"

	blockapi "s3b/vsp-blockchain/p2p-blockchain/blockchain/api"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/block"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/konto"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"s3b/vsp-blockchain/p2p-blockchain/wallet/core/keys"
)

// HistoryAPI provides the interface for querying transaction history.
type HistoryAPI interface {
	// GetHistory returns the transaction history for a given V$Address.
	GetHistory(vsAddress string) konto.HistoryResult
}

// HistoryAPIImpl implements HistoryAPI using the BlockStore.
type HistoryAPIImpl struct {
	blockStore blockapi.BlockStoreVisualizationAPI
	keyDecoder keys.KeyDecoder
}

// NewHistoryAPIImpl creates a new HistoryAPIImpl with the given dependencies.
func NewHistoryAPIImpl(blockStore blockapi.BlockStoreVisualizationAPI, keyDecoder keys.KeyDecoder) *HistoryAPIImpl {
	return &HistoryAPIImpl{
		blockStore: blockStore,
		keyDecoder: keyDecoder,
	}
}

// GetHistory implements HistoryAPI.GetHistory.
func (api *HistoryAPIImpl) GetHistory(vsAddress string) konto.HistoryResult {
	pubKeyHashBytes, result, err := api.validateAddress(vsAddress)
	if err {
		return result
	}

	var pubKeyHash [20]byte
	copy(pubKeyHash[:], pubKeyHashBytes)

	// Find all transactions involving this address in the main chain
	transactions := api.findTransactionsForAddress(pubKeyHash)

	return konto.HistoryResult{
		Success:      true,
		Transactions: transactions,
	}
}

func (api *HistoryAPIImpl) validateAddress(vsAddress string) ([]byte, konto.HistoryResult, bool) {
	// Decode the V$Address to get the public key hash
	pubKeyHashBytes, version, err := api.keyDecoder.Base58CheckToBytes(vsAddress)
	if err != nil {
		return pubKeyHashBytes, konto.HistoryResult{
			Success:      false,
			ErrorMessage: "invalid V$Address format: " + err.Error(),
		}, true
	}

	// V$Address uses version 0x00
	if version != 0x00 {
		return pubKeyHashBytes, konto.HistoryResult{
			Success:      false,
			ErrorMessage: "invalid V$Address version byte",
		}, true
	}

	// Convert to fixed-size array (20 bytes for public key hash)
	if len(pubKeyHashBytes) != 20 {
		return pubKeyHashBytes, konto.HistoryResult{
			Success:      false,
			ErrorMessage: "invalid public key hash length",
		}, true
	}
	return pubKeyHashBytes, konto.HistoryResult{}, false
}

// findTransactionsForAddress finds all main chain transactions involving the given public key hash.
func (api *HistoryAPIImpl) findTransactionsForAddress(pubKeyHash transaction.PubKeyHash) []konto.TransactionEntry {
	var result []konto.TransactionEntry

	// Get all blocks with metadata to identify main chain blocks
	allBlocks := api.blockStore.GetAllBlocksWithMetadata()

	// Build a transaction index for efficient UTXO lookups
	txIndex := api.buildTransactionIndex(allBlocks)

	for _, blockWithMeta := range allBlocks {
		// Only consider main chain blocks
		if !blockWithMeta.IsMainChain {
			continue
		}

		// Check each transaction in the block
		for _, tx := range blockWithMeta.Block.Transactions {
			if api.isAddressInvolved(tx, pubKeyHash) {
				entry := api.createTransactionEntry(tx, blockWithMeta.Height, pubKeyHash, txIndex)
				result = append(result, entry)
			}
		}
	}

	return result
}

// buildTransactionIndex creates a map from TransactionID to Transaction for efficient lookups.
func (api *HistoryAPIImpl) buildTransactionIndex(allBlocks []block.BlockWithMetadata) map[transaction.TransactionID]transaction.Transaction {
	txIndex := make(map[transaction.TransactionID]transaction.Transaction)

	for _, blockWithMeta := range allBlocks {
		// Only index main chain transactions
		if !blockWithMeta.IsMainChain {
			continue
		}

		for _, tx := range blockWithMeta.Block.Transactions {
			txID := tx.TransactionId()
			txIndex[txID] = tx
		}
	}

	return txIndex
}

// isAddressInvolved checks if the given address is involved in the transaction.
func (api *HistoryAPIImpl) isAddressInvolved(tx transaction.Transaction, pubKeyHash transaction.PubKeyHash) bool {
	// Check outputs (receiving)
	for _, output := range tx.Outputs {
		if output.PubKeyHash == pubKeyHash {
			return true
		}
	}

	// Check inputs (sending) - the sender's pubkey hash can be derived from the public key in the input
	for _, input := range tx.Inputs {
		inputPubKeyHash := transaction.Hash160(input.PubKey)
		if inputPubKeyHash == pubKeyHash {
			return true
		}
	}

	return false
}

// createTransactionEntry creates a TransactionEntry from a transaction.
func (api *HistoryAPIImpl) createTransactionEntry(tx transaction.Transaction, blockHeight uint64, pubKeyHash transaction.PubKeyHash, txIndex map[transaction.TransactionID]transaction.Transaction) konto.TransactionEntry {
	var received uint64
	var sent uint64

	// Calculate received amount (from outputs to this address)
	for _, output := range tx.Outputs {
		if output.PubKeyHash == pubKeyHash {
			received += output.Value
		}
	}

	// Calculate sent amount by looking up the referenced UTXOs
	for _, input := range tx.Inputs {
		inputPubKeyHash := transaction.Hash160(input.PubKey)
		if inputPubKeyHash == pubKeyHash {
			// Look up the previous transaction to get the actual value spent
			spentValue := api.lookupPreviousOutputValue(input.PrevTxID, input.OutputIndex, txIndex)
			sent += spentValue
		}
	}

	txID := tx.TransactionId()

	return konto.TransactionEntry{
		TransactionID: hex.EncodeToString(txID[:]),
		BlockHeight:   blockHeight,
		Received:      received,
		Sent:          sent,
		IsSender:      sent > 0,
	}
}

// lookupPreviousOutputValue finds the value of a specific output in a previous transaction.
func (api *HistoryAPIImpl) lookupPreviousOutputValue(prevTxID transaction.TransactionID, outputIndex uint32, txIndex map[transaction.TransactionID]transaction.Transaction) uint64 {
	prevTx, exists := txIndex[prevTxID]
	if !exists {
		return 0
	}

	if int(outputIndex) < len(prevTx.Outputs) {
		return prevTx.Outputs[outputIndex].Value
	}

	return 0
}
