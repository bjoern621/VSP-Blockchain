package core

import (
	"encoding/hex"
	"errors"
	"fmt"
	"s3b/vsp-blockchain/p2p-blockchain/wallet/core/keys"

	blockapi "s3b/vsp-blockchain/p2p-blockchain/blockchain/api"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/inv"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api"

	"bjoernblessin.de/go-utils/util/logger"
)

// TransactionCreationService handles the creation and broadcasting of transactions.
type TransactionCreationService struct {
	keyGenerator  keys.KeyGenerator
	keyDecoder    keys.KeyDecoder
	blockchainAPI api.BlockchainAPI
	utxoAPI       blockapi.UtxoStoreAPI
	blockStore    blockapi.BlockStoreAPI
}

// NewTransactionCreationService creates a new TransactionCreationService with the given dependencies.
func NewTransactionCreationService(
	keyGenerator keys.KeyGenerator,
	keyDecoder keys.KeyDecoder,
	blockchainAPI api.BlockchainAPI,
	utxoAPI blockapi.UtxoStoreAPI,
	blockStore blockapi.BlockStoreAPI,
) *TransactionCreationService {
	return &TransactionCreationService{
		keyGenerator:  keyGenerator,
		keyDecoder:    keyDecoder,
		blockchainAPI: blockchainAPI,
		utxoAPI:       utxoAPI,
		blockStore:    blockStore,
	}
}

// CreateTransaction creates and broadcasts a new transaction.
func (s *TransactionCreationService) CreateTransaction(recipientVSAddress string, amount uint64, senderPrivateKeyWIF string) transaction.TransactionResult {

	keyset, err := s.keyGenerator.GetKeysetFromWIF(senderPrivateKeyWIF)
	if err != nil {
		return s.handleInvalidPrivateKey(err)
	}
	senderPubKeyHash, err := s.decodeVSAddress(keyset.VSAddress)
	if err != nil {
		return s.handleInvalidAddress(err)
	}
	recipientPubKeyHash, err := s.decodeVSAddress(recipientVSAddress)
	if err != nil {
		return s.handleInvalidAddress(err)
	}

	mainChainTip := s.blockStore.GetMainChainTip()
	mainChainTipHash := mainChainTip.Hash()
	utxos, err := s.utxoAPI.GetUtxosByPubKeyHashFromBlock(senderPubKeyHash, mainChainTipHash)
	if err != nil || len(utxos) == 0 {
		return s.handleInsufficientFunds(err)
	}

	privKey := transaction.PrivateKey(keyset.PrivateKey)
	tx, err := transaction.NewTransaction(utxos, recipientPubKeyHash, amount, common.TransactionFee, privKey)
	if err != nil {
		return s.handleTransactionCreationError(err)
	}

	return s.handleSuccess(tx)
}

func (s *TransactionCreationService) handleSuccess(tx *transaction.Transaction) transaction.TransactionResult {
	txID := tx.TransactionId()
	txIDHex := hex.EncodeToString(txID[:])
	invVectors := []*inv.InvVector{
		{
			Hash:    tx.Hash(),
			InvType: inv.InvTypeMsgTx,
		},
	}
	s.blockchainAPI.BroadcastInvExclusionary(invVectors, "") // TODO: Replace with broadcast to all when implemented

	logger.Infof("[wallet] Transaction created and broadcast successfully: %s", txIDHex)

	return transaction.TransactionResult{
		Success:       true,
		ErrorCode:     transaction.ErrorCodeNone,
		ErrorMessage:  "",
		TransactionID: txIDHex,
	}
}

func (s *TransactionCreationService) handleTransactionCreationError(err error) transaction.TransactionResult {
	logger.Warnf("[wallet] Failed to create transaction: %v", err)
	if errors.Is(err, transaction.ErrInsufficientFunds) {
		return s.handleInsufficientFunds(err)
	}
	return transaction.TransactionResult{
		Success:      false,
		ErrorCode:    transaction.ErrorCodeInvalidPrivateKey,
		ErrorMessage: fmt.Sprintf("Failed to create transaction: %v", err),
	}
}

func (s *TransactionCreationService) handleInvalidPrivateKey(err error) transaction.TransactionResult {
	logger.Warnf("[wallet] Failed to decode sender private key WIF: %v", err)
	return transaction.TransactionResult{
		Success:      false,
		ErrorCode:    transaction.ErrorCodeInvalidPrivateKey,
		ErrorMessage: fmt.Sprintf("Invalid sender private key WIF: %v", err),
	}
}

func (s *TransactionCreationService) handleInsufficientFunds(err error) transaction.TransactionResult {
	logger.Warnf("[wallet] Insufficient funds found for sender address")
	return transaction.TransactionResult{
		Success:      false,
		ErrorCode:    transaction.ErrorCodeInsufficientFunds,
		ErrorMessage: fmt.Sprintf("Insufficient funds available for this address, %v", err),
	}
}

func (s *TransactionCreationService) handleInvalidAddress(err error) transaction.TransactionResult {
	logger.Warnf("[wallet] Failed to decode recipient V$Address: %v", err)
	return transaction.TransactionResult{
		Success:      false,
		ErrorCode:    transaction.ErrorCodeValidationFailed,
		ErrorMessage: fmt.Sprintf("Invalid recipient V$Address: %v", err),
	}
}

// decodeVSAddress decodes a V$Address (Base58Check encoded public key hash) to a PubKeyHash.
func (s *TransactionCreationService) decodeVSAddress(vsAddress string) (transaction.PubKeyHash, error) {
	payload, version, err := s.keyDecoder.Base58CheckToBytes(vsAddress)
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
