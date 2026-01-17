package vsgoin_node_adapter

import (
	"context"
	"fmt"
	"s3b/vsp-blockchain/rest-api/internal/common"
	"s3b/vsp-blockchain/rest-api/internal/pb"
)

type TransactionAdapter interface {
	GenerateKeyset() (common.Keyset, error)
	GetKeysetFromWIF(privateKeyWIF string) (common.Keyset, error)
	// CreateTransaction creates and broadcasts a new transaction via the local node.
	CreateTransaction(req common.TransactionRequest) (*common.TransactionResult, error)
	GetBlockchainVisualization() (string, error)
}

type TransactionAdapterImpl struct {
	appServiceClient pb.AppServiceClient
}

func NewTransactionAdapterImpl(appServiceClient pb.AppServiceClient) *TransactionAdapterImpl {

	return &TransactionAdapterImpl{
		appServiceClient: appServiceClient,
	}
}

func (t *TransactionAdapterImpl) GenerateKeyset() (common.Keyset, error) {
	pbKeyset, err := t.appServiceClient.GenerateKeyset(context.Background(), nil)

	if err != nil {
		return common.Keyset{}, common.ErrServer
	}

	return common.Keyset{
		PrivateKey:    [common.PrivateKeySize]byte(pbKeyset.GetKeyset().PrivateKey),
		PrivateKeyWif: pbKeyset.GetKeyset().PrivateKeyWif,
		PublicKey:     [common.PublicKeySize]byte(pbKeyset.GetKeyset().PublicKey),
		VSAddress:     pbKeyset.GetKeyset().VSAddress,
	}, nil
}

func (t *TransactionAdapterImpl) GetKeysetFromWIF(privateKeyWIF string) (common.Keyset, error) {
	request := pb.GetKeysetFromWIFRequest{PrivateKeyWif: privateKeyWIF}
	pbKeyset, err := t.appServiceClient.GetKeysetFromWIF(context.Background(), &request)
	if err != nil {
		return common.Keyset{}, common.ErrServer
	}

	if pbKeyset.FalseInput {
		return common.Keyset{}, common.ErrWIFInput
	}

	return common.Keyset{
		PrivateKey:    [common.PrivateKeySize]byte(pbKeyset.GetKeyset().PrivateKey),
		PrivateKeyWif: pbKeyset.GetKeyset().PrivateKeyWif,
		PublicKey:     [common.PublicKeySize]byte(pbKeyset.GetKeyset().PublicKey),
		VSAddress:     pbKeyset.GetKeyset().VSAddress,
	}, nil
}

// CreateTransaction send transaction request to local node
func (t *TransactionAdapterImpl) CreateTransaction(req common.TransactionRequest) (*common.TransactionResult, error) {

	// Create gRPC request
	grpcReq := &pb.CreateTransactionRequest{
		RecipientVsAddress:  req.RecipientVSAddress,
		Amount:              req.Amount,
		SenderPrivateKeyWif: req.SenderPrivateKeyWIF,
	}

	// Call the gRPC service
	resp, err := t.appServiceClient.CreateTransaction(context.Background(), grpcReq)
	if err != nil {
		return nil, fmt.Errorf("gRPC call failed: %w", err)
	}

	// Map gRPC response to adapter result
	return &common.TransactionResult{
		Success:       resp.Success,
		ErrorCode:     mapErrorCode(resp.ErrorCode),
		ErrorMessage:  resp.ErrorMessage,
		TransactionID: resp.TransactionId,
	}, nil
}

// mapErrorCode converts gRPC error codes to adapter error codes.
func mapErrorCode(code pb.TransactionErrorCode) common.TransactionErrorCode {
	switch code {
	case pb.TransactionErrorCode_NONE:
		return common.ErrorCodeNone
	case pb.TransactionErrorCode_INVALID_PRIVATE_KEY:
		return common.ErrorCodeInvalidPrivateKey
	case pb.TransactionErrorCode_INSUFFICIENT_FUNDS:
		return common.ErrorCodeInsufficientFunds
	case pb.TransactionErrorCode_VALIDATION_FAILED:
		return common.ErrorCodeValidationFailed
	case pb.TransactionErrorCode_BROADCAST_FAILED:
		return common.ErrorCodeBroadcastFailed
	default:
		return common.ErrorCodeValidationFailed
	}
}

func (t *TransactionAdapterImpl) GetBlockchainVisualization() (string, error) {
	result, err := t.appServiceClient.GetBlockchainVisualization(context.Background(), nil)
	if err != nil {
		return "", common.ErrServer
	}

	return result.VisualizationUrl, nil
}
