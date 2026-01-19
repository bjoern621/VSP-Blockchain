package transaktionsverlauf

import (
	"s3b/vsp-blockchain/rest-api/internal/common"
	"s3b/vsp-blockchain/rest-api/vsgoin_node_adapter"
)

// TransaktionsverlaufService provides the interface for querying transaction history.
type TransaktionsverlaufService struct {
	historyAdapter vsgoin_node_adapter.HistoryAdapterAPI
}

// NewTransaktionsverlaufService creates a new TransaktionsverlaufService with the given adapter.
func NewTransaktionsverlaufService(historyAdapter vsgoin_node_adapter.HistoryAdapterAPI) *TransaktionsverlaufService {
	return &TransaktionsverlaufService{
		historyAdapter: historyAdapter,
	}
}

// GetHistory returns the transaction history for the given V$Address.
func (s *TransaktionsverlaufService) GetHistory(vsAddress string) ([]string, error) {
	// Validate the V$Address format
	if vsAddress == "" {
		return nil, common.ErrInvalidAddress
	}

	// Query the local node for transaction history
	result, err := s.historyAdapter.GetHistory(vsAddress)
	if err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, &common.AssetError{Message: result.ErrorMessage}
	}

	return result.Transactions, nil
}
