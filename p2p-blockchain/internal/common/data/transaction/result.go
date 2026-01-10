package transaction

// TransactionResult contains the result of a transaction creation attempt.
type TransactionResult struct {
	Success       bool
	ErrorCode     TransactionErrorCode
	ErrorMessage  string
	TransactionID string
}

// TransactionErrorCode categorizes transaction creation failures.
type TransactionErrorCode int

const (
	ErrorCodeNone TransactionErrorCode = iota
	ErrorCodeInvalidPrivateKey
	ErrorCodeInsufficientFunds
	ErrorCodeValidationFailed
	ErrorCodeBroadcastFailed
)
