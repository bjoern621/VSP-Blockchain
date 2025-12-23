package core

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/inv"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api"
	walletapi "s3b/vsp-blockchain/p2p-blockchain/wallet/api"

	"bjoernblessin.de/go-utils/util/logger"
	"github.com/akamensky/base58"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// TransactionErrorCode categorizes transaction creation failures.
type TransactionErrorCode int

const (
	ErrorCodeNone TransactionErrorCode = iota
	ErrorCodeInvalidPrivateKey
	ErrorCodeInsufficientFunds
	ErrorCodeValidationFailed
	ErrorCodeBroadcastFailed
)

// TransactionResult contains the result of a transaction creation attempt.
type TransactionResult struct {
	Success       bool
	ErrorCode     TransactionErrorCode
	ErrorMessage  string
	TransactionID string
}

// TransactionService handles the creation and broadcasting of transactions.
type TransactionService struct {
	keyGeneratorAPI walletapi.KeyGeneratorApi
	blockchainAPI   api.BlockchainAPI
}

// NewTransactionService creates a new TransactionService with the given dependencies.
func NewTransactionService(
	keyGeneratorAPI walletapi.KeyGeneratorApi,
	blockchainAPI api.BlockchainAPI,
) *TransactionService {
	return &TransactionService{
		keyGeneratorAPI: keyGeneratorAPI,
		blockchainAPI:   blockchainAPI,
	}
}

// CreateTransaction creates and broadcasts a new transaction.
func (s *TransactionService) CreateTransaction(recipientVSAddress string, amount uint64, senderPrivateKeyWIF string) TransactionResult {
	// 1. Validate and decode the sender's private key from WIF
	keyset, err := s.keyGeneratorAPI.GetKeysetFromWIF(senderPrivateKeyWIF)
	if err != nil {
		logger.Warnf("Invalid private key WIF: %v", err)
		return TransactionResult{
			Success:      false,
			ErrorCode:    ErrorCodeInvalidPrivateKey,
			ErrorMessage: fmt.Sprintf("Invalid private key: %v", err),
		}
	}

	// 2. Convert private key bytes to ecdsa.PrivateKey
	privKey := secp256k1.PrivKeyFromBytes(keyset.PrivateKey[:])
	ecdsaPrivKey := privKey.ToECDSA()

	// 3. Decode the recipient's V$Address to get the public key hash
	recipientPubKeyHash, err := decodeVSAddress(recipientVSAddress)
	if err != nil {
		logger.Warnf("Invalid recipient V$Address: %v", err)
		return TransactionResult{
			Success:      false,
			ErrorCode:    ErrorCodeValidationFailed,
			ErrorMessage: fmt.Sprintf("Invalid recipient address: %v", err),
		}
	}

	// 4. Get the sender's public key hash from keyset
	senderPubKeyHash := transaction.Hash160(transaction.PubKey(keyset.PublicKey))

	// TODO: 5. Get UTXOs for the sender's address
	utxos := []transaction.UTXO{}
	if len(utxos) == 0 {
		logger.Warnf("No UTXOs found for sender address")
		return TransactionResult{
			Success:      false,
			ErrorCode:    ErrorCodeInsufficientFunds,
			ErrorMessage: "No funds available for this address",
		}
	}

	// 6. Create the transaction (includes UTXO selection and signing)
	// Using a fixed fee of 1 for now
	const transactionFee uint64 = 1
	tx, err := transaction.NewTransaction(utxos, recipientPubKeyHash, amount, transactionFee, ecdsaPrivKey)
	if err != nil {
		logger.Warnf("Failed to create transaction: %v", err)
		if err.Error() == "insufficient funds" {
			return TransactionResult{
				Success:      false,
				ErrorCode:    ErrorCodeInsufficientFunds,
				ErrorMessage: "Insufficient funds to complete the transaction",
			}
		}
		return TransactionResult{
			Success:      false,
			ErrorCode:    ErrorCodeValidationFailed,
			ErrorMessage: fmt.Sprintf("Failed to create transaction: %v", err),
		}
	}

	// 7. Get the transaction ID
	txID := tx.TransactionId()
	txIDHex := hex.EncodeToString(txID[:])

	// 8. Broadcast the transaction to the network
	// Create inventory vector for the transaction
	invVectors := []*inv.InvVector{
		{
			Hash:    tx.Hash(),
			InvType: inv.InvTypeMsgTx,
		},
	}

	// Broadcast to all peers (using empty peer ID to indicate broadcast to all)
	s.blockchainAPI.BroadcastInvExclusionary(invVectors, common.PeerId(""))

	logger.Infof("Transaction created and broadcast successfully: %s", txIDHex)

	return TransactionResult{
		Success:       true,
		ErrorCode:     ErrorCodeNone,
		ErrorMessage:  "",
		TransactionID: txIDHex,
	}
}

// decodeVSAddress decodes a V$Address (Base58Check encoded public key hash) to a PubKeyHash.
func decodeVSAddress(vsAddress string) (transaction.PubKeyHash, error) {
	// V$Address is Base58Check encoded with version byte 0x00
	// We need to decode it and extract the 20-byte public key hash
	payload, version, err := base58CheckDecode(vsAddress)
	if err != nil {
		return transaction.PubKeyHash{}, fmt.Errorf("failed to decode V$Address: %w", err)
	}

	// Version byte for V$Address should be 0x00
	if version != 0x00 {
		return transaction.PubKeyHash{}, fmt.Errorf("invalid V$Address version: expected 0x00, got 0x%02x", version)
	}

	if len(payload) != 20 {
		return transaction.PubKeyHash{}, fmt.Errorf("invalid V$Address: expected 20 bytes, got %d", len(payload))
	}

	var pubKeyHash transaction.PubKeyHash
	copy(pubKeyHash[:], payload)
	return pubKeyHash, nil
}

// base58CheckDecode decodes a Base58Check encoded string and returns the payload and version.
func base58CheckDecode(input string) ([]byte, byte, error) {
	bytes, err := base58.Decode(input)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to decode base58: %w", err)
	}

	if len(bytes) < 5 {
		return nil, 0, fmt.Errorf("invalid base58check: too short")
	}

	version := bytes[0]
	payload := bytes[1 : len(bytes)-4]
	checksumBytes := bytes[len(bytes)-4:]

	// Verify checksum
	expectedChecksum := getFirstFourChecksumBytes([]byte{version}, payload)
	for i := 0; i < 4; i++ {
		if checksumBytes[i] != expectedChecksum[i] {
			return nil, 0, fmt.Errorf("invalid base58check: checksum mismatch")
		}
	}

	return payload, version, nil
}

// getFirstFourChecksumBytes computes the first four bytes of the double SHA256 checksum.
func getFirstFourChecksumBytes(parts ...[]byte) []byte {
	h := sha256.New()
	for _, part := range parts {
		h.Write(part)
	}
	firstHash := h.Sum(nil)
	secondHash := sha256.Sum256(firstHash)
	return secondHash[:4]
}
