package common

// TransactionErrorCode categorizes transaction creation failures.
type TransactionErrorCode int

const (
	ErrorCodeNone TransactionErrorCode = iota
	ErrorCodeInvalidPrivateKey
	ErrorCodeInsufficientFunds
	ErrorCodeValidationFailed
	ErrorCodeBroadcastFailed
)
