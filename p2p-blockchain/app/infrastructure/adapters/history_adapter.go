package adapters

import (
	"fmt"
	appapi "s3b/vsp-blockchain/p2p-blockchain/app/api"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/konto"
	"s3b/vsp-blockchain/p2p-blockchain/internal/pb"

	"bjoernblessin.de/go-utils/util/logger"
)

// HistoryHandlerAdapter handles history queries from gRPC requests.
type HistoryHandlerAdapter struct {
	historyAPI appapi.HistoryAPI
}

// NewHistoryAdapter creates a new HistoryHandlerAdapter with the given history API.
func NewHistoryAdapter(api appapi.HistoryAPI) *HistoryHandlerAdapter {
	return &HistoryHandlerAdapter{
		historyAPI: api,
	}
}

// GetHistory handles the GetHistory RPC call from external local systems.
func (h *HistoryHandlerAdapter) GetHistory(req *pb.GetHistoryRequest) *pb.GetHistoryResponse {
	if req.VsAddress == "" {
		return &pb.GetHistoryResponse{
			Success:      false,
			ErrorMessage: "V$Address is required",
		}
	}

	result := h.historyAPI.GetHistory(req.VsAddress)

	if !result.Success {
		return &pb.GetHistoryResponse{
			Success:      false,
			ErrorMessage: result.ErrorMessage,
		}
	}

	// Convert transactions to string format for response
	txStrings := make([]string, 0, len(result.Transactions))
	for _, tx := range result.Transactions {
		txStr := h.formatTxString(tx)
		txStrings = append(txStrings, txStr)
	}

	return &pb.GetHistoryResponse{
		Success:      true,
		Transactions: txStrings,
	}
}

func (h *HistoryHandlerAdapter) formatTxString(tx konto.TransactionEntry) string {
	var txStr string
	if tx.IsSender && tx.Received > 0 {
		// Both sent and received (e.g., change back to self)
		txStr = fmt.Sprintf("TxID: %s, Block: %d, Sent: %d, Received: %d",
			tx.TransactionID, tx.BlockHeight, tx.Sent, tx.Received)
	} else if tx.IsSender {
		// Only sent
		txStr = fmt.Sprintf("TxID: %s, Block: %d, Sent: %d",
			tx.TransactionID, tx.BlockHeight, tx.Sent)
	} else if tx.Received > 0 {
		// Only received
		txStr = fmt.Sprintf("TxID: %s, Block: %d, Received: %d",
			tx.TransactionID, tx.BlockHeight, tx.Received)
	} else {
		logger.Warnf("TransactionEntry with zero sent and received amounts: %+v", tx)
	}
	return txStr
}
