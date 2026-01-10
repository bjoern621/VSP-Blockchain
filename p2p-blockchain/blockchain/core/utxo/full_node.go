package utxo

import (
	"errors"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/data/utxopool"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
)

var (
	ErrUTXOAlreadySpent = errors.New("UTXO already spent")
	ErrUTXONotFound     = errors.New("UTXO not found")
)

// FullNodeUTXOService provides a unified view of both confirmed (chainstate) and
// unconfirmed (mempool) UTXOs. It implements the LookupService interface
// for use in transaction validation.
//
// Lookup order:
// 1. Check if spent in mempool -> not available
// 2. Check mempool UTXOs -> return if found
// 3. Check chainstate -> return if found
type FullNodeUTXOService struct {
	mempool    *MemUTXOPoolService
	chainstate *ChainStateService
}

var _ UTXOService = (*FullNodeUTXOService)(nil)

// NewFullNodeUTXOService creates a new combined UTXO pool
func NewFullNodeUTXOService(mempool *MemUTXOPoolService, chainstate *ChainStateService) *FullNodeUTXOService {
	return &FullNodeUTXOService{
		mempool:    mempool,
		chainstate: chainstate,
	}
}

// GetUTXO retrieves an output by transaction ID and output index
func (c *FullNodeUTXOService) GetUTXO(txID transaction.TransactionID, outputIndex uint32) (transaction.Output, error) {
	outpoint := utxopool.NewOutpoint(txID, outputIndex)

	if c.mempool.IsSpent(outpoint) {
		return transaction.Output{}, ErrUTXOAlreadySpent
	}

	if entry, err := c.mempool.Get(outpoint); err == nil {
		return entry.Output, nil
	}

	entry, err := c.chainstate.Get(outpoint)
	if err == nil {
		return entry.Output, nil
	}

	return transaction.Output{}, err
}

// GetUTXOEntry retrieves the full UTXO entry with metadata
func (c *FullNodeUTXOService) GetUTXOEntry(outpoint utxopool.Outpoint) (utxopool.UTXOEntry, error) {
	if c.mempool.IsSpent(outpoint) {
		return utxopool.UTXOEntry{}, ErrUTXOAlreadySpent
	}

	if entry, err := c.mempool.Get(outpoint); err == nil {
		return entry, nil
	}

	return c.chainstate.Get(outpoint)
}

// ContainsUTXO checks if a UTXO exists and is unspent
func (c *FullNodeUTXOService) ContainsUTXO(outpoint utxopool.Outpoint) bool {
	if c.mempool.IsSpent(outpoint) {
		return false
	}
	if c.mempool.Contains(outpoint) {
		return true
	}
	return c.chainstate.Contains(outpoint)
}

func (c *FullNodeUTXOService) Remove(outpoint utxopool.Outpoint) error {
	err := c.mempool.Remove(outpoint)
	if err != nil {
		return err
	}
	return c.chainstate.Remove(outpoint)
}

// AddUTXO adds a new UTXO to the appropriate pool
// If blockHeight > 0, it goes to chainstate, otherwise to mempool
func (c *FullNodeUTXOService) AddUTXO(outpoint utxopool.Outpoint, entry utxopool.UTXOEntry) error {
	if entry.IsConfirmed() {
		return c.chainstate.Add(outpoint, entry)
	}
	return c.mempool.Add(outpoint, entry)
}

// SpendUTXO marks a UTXO as spent
func (c *FullNodeUTXOService) SpendUTXO(outpoint utxopool.Outpoint) error {
	// Check mempool first
	if c.mempool.Contains(outpoint) {
		return c.mempool.Remove(outpoint)
	}

	// Check if in chainstate
	if c.chainstate.Contains(outpoint) {
		// Mark as spent in mempool (don't remove from chainstate yet)
		// It will be removed when the spending transaction is confirmed
		c.mempool.MarkSpent(outpoint)
		return nil
	}

	return ErrUTXONotFound
}

// ApplyTransaction applies a transaction to the UTXO set
// For unconfirmed transactions (blockHeight <= 5), changes go to mempool
// For confirmed transactions (blockHeight > 5), changes go to chainstate
func (c *FullNodeUTXOService) ApplyTransaction(
	tx *transaction.Transaction,
	txID transaction.TransactionID,
	blockHeight uint64,
	isCoinbase bool,
) error {
	isConfirmed := blockHeight > 5

	// Spend inputs
	for _, input := range tx.Inputs {
		outpoint := utxopool.NewOutpoint(input.PrevTxID, input.OutputIndex)

		if isConfirmed {
			// For confirmed transactions:
			// 1. Remove from chainstate
			// 2. Remove spent marker from mempool
			if err := c.chainstate.Remove(outpoint); err != nil {
				return err
			}
			c.mempool.UnmarkSpent(outpoint)
			// Also remove from mempool if it was an unconfirmed UTXO
			_ = c.mempool.Remove(outpoint)
		} else {
			// For unconfirmed transactions, mark as spent
			c.mempool.MarkSpent(outpoint)
		}
	}

	// Add outputs
	for i, output := range tx.Outputs {
		outpoint := utxopool.NewOutpoint(txID, uint32(i))
		entry := utxopool.NewUTXOEntry(output, blockHeight, isCoinbase && i == 0)

		if isConfirmed {
			if err := c.chainstate.Add(outpoint, entry); err != nil {
				return err
			}
			// Remove from mempool if it existed there (transaction was confirmed)
			_ = c.mempool.Remove(outpoint)
		} else {
			if err := c.mempool.Add(outpoint, entry); err != nil {
				return err
			}
		}
	}

	return nil
}

// RevertTransaction reverts a transaction from the UTXO set
// This is used during chain reorganizations
func (c *FullNodeUTXOService) RevertTransaction(
	tx *transaction.Transaction,
	txID transaction.TransactionID,
	inputUTXOs []utxopool.UTXOEntry,
) error {
	// Remove outputs (they are no longer valid)
	for i := range tx.Outputs {
		outpoint := utxopool.NewOutpoint(txID, uint32(i))
		if err := c.chainstate.Remove(outpoint); err != nil {
			return err
		}
	}

	// Restore inputs (they are now unspent again)
	for i, input := range tx.Inputs {
		outpoint := utxopool.NewOutpoint(input.PrevTxID, input.OutputIndex)
		if i < len(inputUTXOs) {
			if err := c.chainstate.Add(outpoint, inputUTXOs[i]); err != nil {
				return err
			}
		}
	}

	return nil
}

// Flush persists any pending changes to disk
func (c *FullNodeUTXOService) Flush() error {
	return c.chainstate.Flush()
}

// Close closes both pools and releases resources
func (c *FullNodeUTXOService) Close() error {
	if err := c.mempool.Close(); err != nil {
		return err
	}
	return c.chainstate.Close()
}

// ClearMempool clears the mempool UTXO set
func (c *FullNodeUTXOService) ClearMempool() {
	c.mempool.Clear()
}

// GetMempool returns the mempool for direct access
func (c *FullNodeUTXOService) GetMempool() *MemUTXOPoolService {
	return c.mempool
}

// GetChainState returns the chainstate for direct access
func (c *FullNodeUTXOService) GetChainState() *ChainStateService {
	return c.chainstate
}

// GetUTXOsByPubKeyHash returns all UTXOs (mempool + chainstate) for a given PubKeyHash.
// It excludes outpoints that are marked as spent in the mempool.
// Results are returned in no particular order.
func (c *FullNodeUTXOService) GetUTXOsByPubKeyHash(pubKeyHash transaction.PubKeyHash) ([]transaction.UTXO, error) {
	// Get spent markers from mempool to exclude from chainstate results
	spentOutpoints := c.mempool.GetSpentOutpoints()

	results := make([]transaction.UTXO, 0)

	// 1. Get unconfirmed UTXOs from mempool
	mempoolEntries, err := c.mempool.GetUTXOsWithOutpointByPubKeyHash(pubKeyHash)
	if err != nil {
		return nil, err
	}
	results = append(results, mempoolEntries...)

	// 2. Get chainstate UTXOs
	chainstateEntries, err := c.chainstate.GetUTXOsByPubKeyHash(pubKeyHash)
	if err != nil {
		return nil, err
	}

	for _, uwp := range chainstateEntries {
		// Check if this UTXO is spent in mempool
		if _, spent := spentOutpoints[string(utxopool.NewOutpoint(uwp.TxID, uwp.OutputIndex).Key())]; spent {
			continue
		}
		results = append(results, uwp)
	}

	return results, nil
}
