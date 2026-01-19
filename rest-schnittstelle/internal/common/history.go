package common

// HistoryResult represents the outcome of a transaction history query.
type HistoryResult struct {
	Success      bool
	ErrorMessage string
	Transactions []string
}
