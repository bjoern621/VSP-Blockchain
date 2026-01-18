package konto

// TransactionEntry represents a single transaction in the history.
type TransactionEntry struct {
	// TransactionID is the unique identifier of the transaction.
	TransactionID string
	// BlockHeight is the height of the block containing this transaction.
	BlockHeight uint64
	// Received is the amount received by the queried address in this transaction.
	Received uint64
	// Sent is the total amount sent from the queried address in this transaction (sum of spent UTXOs).
	Sent uint64
	// IsSender indicates if the queried address was a sender in this transaction.
	IsSender bool
}

// HistoryResult represents the outcome of a transaction history query.
type HistoryResult struct {
	Success      bool
	ErrorMessage string
	Transactions []TransactionEntry
}
