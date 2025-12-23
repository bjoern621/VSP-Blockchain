// Package core contains the core business logic services for the app layer.
package core

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/utxo"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"s3b/vsp-blockchain/p2p-blockchain/wallet/core/keys"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// TransactionResult represents the outcome of a transaction creation.
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
	ErrorCodeInvalidRecipient
	ErrorCodeValidationFailed
	ErrorCodeBroadcastFailed
)

// TransactionService handles transaction creation and broadcasting.
type TransactionService struct {
	keyGenerator keys.KeyGenerator
	keyDecoder   keys.KeyDecoder
	utxoService  utxo.UTXOLookupService
	mempool      *core.Mempool
	// broadcaster to send transactions to the network would be added here
}

// NewTransactionService creates a new TransactionService.
func NewTransactionService(
	keyGenerator keys.KeyGenerator,
	keyDecoder keys.KeyDecoder,
	utxoService utxo.UTXOLookupService,
	mempool *core.Mempool,
) *TransactionService {
	return &TransactionService{
		keyGenerator: keyGenerator,
		keyDecoder:   keyDecoder,
		utxoService:  utxoService,
		mempool:      mempool,
	}
}

// CreateTransaction creates a new transaction from the given parameters.
func (s *TransactionService) CreateTransaction(
	recipientVSAddress string,
	amount uint64,
	senderPrivateKeyWIF string,
) *TransactionResult {
	// 1. Validate and decode the private key
	keyset, err := s.keyGenerator.GetKeysetFromWIF(senderPrivateKeyWIF)
	if err != nil {
		return &TransactionResult{
			Success:      false,
			ErrorCode:    ErrorCodeInvalidPrivateKey,
			ErrorMessage: fmt.Sprintf("Invalid private key: %v", err),
		}
	}

	// 2. Validate and decode the recipient address
	recipientPubKeyHash, err := s.decodeVSAddress(recipientVSAddress)
	if err != nil {
		return &TransactionResult{
			Success:      false,
			ErrorCode:    ErrorCodeInvalidRecipient,
			ErrorMessage: fmt.Sprintf("Invalid recipient address: %v", err),
		}
	}

	// 3. Get UTXOs for the sender
	senderPubKeyHash := transaction.Hash160(transaction.PubKey(keyset.PublicKey))
	utxos, err := s.getUTXOsForAddress(senderPubKeyHash)
	if err != nil {
		return &TransactionResult{
			Success:      false,
			ErrorCode:    ErrorCodeInsufficientFunds,
			ErrorMessage: fmt.Sprintf("Failed to get UTXOs: %v", err),
		}
	}

	// 4. Create ECDSA private key for signing
	privateKey, err := s.createECDSAPrivateKey(keyset.PrivateKey)
	if err != nil {
		return &TransactionResult{
			Success:      false,
			ErrorCode:    ErrorCodeInvalidPrivateKey,
			ErrorMessage: fmt.Sprintf("Failed to create signing key: %v", err),
		}
	}

	// 5. Create the transaction (uses default fee of 0 for now)
	tx, err := transaction.NewTransaction(utxos, recipientPubKeyHash, amount, 0, privateKey)
	if err != nil {
		if errors.Is(err, errors.New("insufficient funds")) {
			return &TransactionResult{
				Success:      false,
				ErrorCode:    ErrorCodeInsufficientFunds,
				ErrorMessage: err.Error(),
			}
		}
		return &TransactionResult{
			Success:      false,
			ErrorCode:    ErrorCodeValidationFailed,
			ErrorMessage: fmt.Sprintf("Failed to create transaction: %v", err),
		}
	}

	// 6. Add to mempool (validates and stores)
	isNew := s.mempool.AddTransaction(*tx)
	if !isNew {
		return &TransactionResult{
			Success:      false,
			ErrorCode:    ErrorCodeValidationFailed,
			ErrorMessage: "Transaction already exists in mempool",
		}
	}

	// 7. Get transaction ID
	txID := tx.TransactionId()
	txIDHex := hex.EncodeToString(txID[:])

	// TODO: Broadcast transaction to the network
	// This would involve sending the transaction to connected peers

	return &TransactionResult{
		Success:       true,
		ErrorCode:     ErrorCodeNone,
		ErrorMessage:  "",
		TransactionID: txIDHex,
	}
}

// decodeVSAddress decodes a V$Address (Base58Check with 0x00 prefix) to a PubKeyHash.
func (s *TransactionService) decodeVSAddress(vsAddress string) (transaction.PubKeyHash, error) {
	keyEncodings := keys.NewKeyEncodingsImpl()
	payload, version, err := keyEncodings.Base58CheckToBytes(vsAddress)
	if err != nil {
		return transaction.PubKeyHash{}, fmt.Errorf("failed to decode address: %w", err)
	}

	if version != 0x00 {
		return transaction.PubKeyHash{}, fmt.Errorf("invalid address version: expected 0x00, got 0x%02x", version)
	}

	if len(payload) != 20 {
		return transaction.PubKeyHash{}, fmt.Errorf("invalid address length: expected 20, got %d", len(payload))
	}

	var pubKeyHash transaction.PubKeyHash
	copy(pubKeyHash[:], payload)
	return pubKeyHash, nil
}

// getUTXOsForAddress retrieves all UTXOs for a given address.
// TODO: This needs to be implemented with a proper UTXO index by address.
func (s *TransactionService) getUTXOsForAddress(pubKeyHash transaction.PubKeyHash) ([]transaction.UTXO, error) {
	// This is a placeholder - in a real implementation, we would query the UTXO set
	// for all UTXOs belonging to this address.
	// For now, return empty slice which will result in insufficient funds.
	return []transaction.UTXO{}, nil
}

// createECDSAPrivateKey creates an ECDSA private key from raw bytes.
func (s *TransactionService) createECDSAPrivateKey(privateKeyBytes [32]byte) (*ecdsa.PrivateKey, error) {
	curve := secp256k1.S256()

	d := new(big.Int).SetBytes(privateKeyBytes[:])

	// Compute public key: Q = d * G
	x, y := curve.ScalarBaseMult(privateKeyBytes[:])

	return &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: curve,
			X:     x,
			Y:     y,
		},
		D: d,
	}, nil
}
