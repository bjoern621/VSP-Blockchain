package api

import (
	"s3b/vsp-blockchain/p2p-blockchain/app/core"
	common "s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
)

// TransactionAPI useless API that wraps wallet's TransactionCreationAPI
type TransactionAPI interface {
	// CreateTransaction creates and broadcasts a new transaction.
	// It validates the private key, checks for sufficient funds, creates the transaction,
	// and broadcasts it to the network.
	//
	// Parameters:
	//   - recipientVSAddress: The recipient's V$Address (Base58Check encoded public key hash)
	//   - amount: The amount of V$Goin to transfer (must be >= 1)
	//   - senderPrivateKeyWIF: The sender's private key in WIF format (Base58Check encoded)
	//
	// Returns:
	//   - TransactionResult containing success status, transaction ID, and any error details
	CreateTransaction(recipientVSAddress string, amount uint64, senderPrivateKeyWIF string) common.TransactionResult
}

// TransactionAPIImpl implements TransactionAPI using the core TransactionService.
type TransactionAPIImpl struct {
	transactionService *core.TransactionService
}

// NewTransactionAPIImpl creates a new TransactionAPIImpl with the given dependencies.
func NewTransactionAPIImpl(transactionService *core.TransactionService) *TransactionAPIImpl {
	return &TransactionAPIImpl{
		transactionService: transactionService,
	}
}

// CreateTransaction implements TransactionAPI.CreateTransaction.
func (api *TransactionAPIImpl) CreateTransaction(recipientVSAddress string, amount uint64, senderPrivateKeyWIF string) common.TransactionResult {
	return api.transactionService.CreateTransaction(recipientVSAddress, amount, senderPrivateKeyWIF)
}
