package common

// TransactionResult represents the outcome of a transaction creation request.
type TransactionResult struct {
	Success       bool
	ErrorCode     TransactionErrorCode
	ErrorMessage  string
	TransactionID string
}

// TransactionRequest contains the data needed to create a new transaction.
type TransactionRequest struct {
	RecipientVSAddress  string
	Amount              uint64
	SenderPrivateKeyWIF string
}
