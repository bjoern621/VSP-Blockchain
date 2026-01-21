package api

import (
	common "s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"s3b/vsp-blockchain/p2p-blockchain/wallet/core"
)

// TransactionCreationAPI API for creating transactions.
// Part of WalletAppAPI.
type TransactionCreationAPI interface {
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

// TransactionCreationAPIImpl implements TransactionCreationAPI using the core TransactionCreationService.
type TransactionCreationAPIImpl struct {
	transactionService *core.TransactionCreationService
}

// NewTransactionCreationAPIImpl creates a new TransactionCreationAPIImpl with the given dependencies.
func NewTransactionCreationAPIImpl(transactionService *core.TransactionCreationService) *TransactionCreationAPIImpl {
	return &TransactionCreationAPIImpl{
		transactionService: transactionService,
	}
}

// CreateTransaction implements TransactionCreationAPI.CreateTransaction.
func (api *TransactionCreationAPIImpl) CreateTransaction(recipientVSAddress string, amount uint64, senderPrivateKeyWIF string) common.TransactionResult {
	return api.transactionService.CreateTransaction(recipientVSAddress, amount, senderPrivateKeyWIF)
}
