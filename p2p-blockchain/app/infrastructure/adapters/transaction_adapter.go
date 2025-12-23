package adapters

import (
	appapi "s3b/vsp-blockchain/p2p-blockchain/app/api"
	appcore "s3b/vsp-blockchain/p2p-blockchain/app/core"
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"
)

type TransactionHandlerAdapter struct {
	transactionAPI appapi.TransactionAPI
}

func NewTransactionAdapter(api appapi.TransactionAPI) *TransactionHandlerAdapter {
	return &TransactionHandlerAdapter{
		transactionAPI: api,
	}
}

// CreateTransaction handles the CreateTransaction RPC call from external local systems.
// It creates and broadcasts a new transaction to the network.
func (t *TransactionHandlerAdapter) CreateTransaction(req *pb.CreateTransactionRequest) *pb.CreateTransactionResponse {
	response := t.validateFields(req)
	if response != nil {
		return response
	}

	result := t.transactionAPI.CreateTransaction(
		req.RecipientVsAddress,
		req.Amount,
		req.SenderPrivateKeyWif,
	)

	pbErrorCode := t.mapErrorCode(result)

	return &pb.CreateTransactionResponse{
		Success:       result.Success,
		ErrorCode:     pbErrorCode,
		ErrorMessage:  result.ErrorMessage,
		TransactionId: result.TransactionID,
	}
}

func (t *TransactionHandlerAdapter) mapErrorCode(result appcore.TransactionResult) pb.TransactionErrorCode {
	var pbErrorCode pb.TransactionErrorCode
	switch result.ErrorCode {
	case appcore.ErrorCodeNone:
		pbErrorCode = pb.TransactionErrorCode_NONE
	case appcore.ErrorCodeInvalidPrivateKey:
		pbErrorCode = pb.TransactionErrorCode_INVALID_PRIVATE_KEY
	case appcore.ErrorCodeInsufficientFunds:
		pbErrorCode = pb.TransactionErrorCode_INSUFFICIENT_FUNDS
	case appcore.ErrorCodeValidationFailed:
		pbErrorCode = pb.TransactionErrorCode_VALIDATION_FAILED
	case appcore.ErrorCodeBroadcastFailed:
		pbErrorCode = pb.TransactionErrorCode_BROADCAST_FAILED
	default:
		pbErrorCode = pb.TransactionErrorCode_VALIDATION_FAILED
	}
	return pbErrorCode
}

func (t *TransactionHandlerAdapter) validateFields(req *pb.CreateTransactionRequest) *pb.CreateTransactionResponse {
	if req.RecipientVsAddress == "" {
		return &pb.CreateTransactionResponse{
			Success:      false,
			ErrorCode:    pb.TransactionErrorCode_VALIDATION_FAILED,
			ErrorMessage: "recipient address is required",
		}
	}

	if req.Amount == 0 {
		return &pb.CreateTransactionResponse{
			Success:      false,
			ErrorCode:    pb.TransactionErrorCode_VALIDATION_FAILED,
			ErrorMessage: "amount must be greater than 0",
		}
	}

	if req.SenderPrivateKeyWif == "" {
		return &pb.CreateTransactionResponse{
			Success:      false,
			ErrorCode:    pb.TransactionErrorCode_INVALID_PRIVATE_KEY,
			ErrorMessage: "sender private key is required",
		}
	}
	return nil
}
