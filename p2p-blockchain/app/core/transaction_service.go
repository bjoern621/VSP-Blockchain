// Package core contains the core business logic services for the app layer.
package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	walletapi "s3b/vsp-blockchain/p2p-blockchain/wallet/api"
)

// TransactionService handles the creation and broadcasting of transactions.
type TransactionService struct {
	transactionCreationAPI walletapi.TransactionCreationAPI
}

// NewTransactionService creates a new TransactionService with the given dependencies.
func NewTransactionService(transactionCreationAPI walletapi.TransactionCreationAPI) *TransactionService {
	return &TransactionService{
		transactionCreationAPI: transactionCreationAPI,
	}
}

// CreateTransaction creates and broadcasts a new transaction.
func (s *TransactionService) CreateTransaction(recipientVSAddress string, amount uint64, senderPrivateKeyWIF string) transaction.TransactionResult {
	return s.transactionCreationAPI.CreateTransaction(recipientVSAddress, amount, senderPrivateKeyWIF)
}
