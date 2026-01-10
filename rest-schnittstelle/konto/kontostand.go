// Package konto contains domain logic for balance/kontostand operations.
package konto

import (
	"s3b/vsp-blockchain/rest-api/internal/common"
	"s3b/vsp-blockchain/rest-api/vsgoin_node_adapter"
)

// KontostandService handles balance/kontostand domain logic.
type KontostandService struct {
	kontoAdapter vsgoin_node_adapter.KontoAdapterAPI
}

// NewKontostandService creates a new KontostandService with the given adapter.
func NewKontostandService(kontoAdapter vsgoin_node_adapter.KontoAdapterAPI) *KontostandService {
	return &KontostandService{
		kontoAdapter: kontoAdapter,
	}
}

// GetBalance retrieves the balance for the given VSAddress.
// Returns a ValidationError if validation fails.
func (s *KontostandService) GetBalance(vsAddress string) (uint64, error) {
	if validationErr := s.validateAddress(vsAddress); validationErr != nil {
		return 0, validationErr
	}

	result, err := s.kontoAdapter.GetAssets(vsAddress)
	if err != nil {
		return 0, err
	}

	if !result.Success {
		return 0, &common.AssetError{Message: result.ErrorMessage}
	}

	// Calculate total balance
	var balance uint64
	for _, asset := range result.Assets {
		balance += asset.Value
	}

	return balance, nil
}

// validateAddress validates the VSAddress format.
func (s *KontostandService) validateAddress(vsAddress string) error {
	if vsAddress == "" || !common.VsAddressPattern.MatchString(vsAddress) {
		return common.ErrInvalidAddress
	}

	return nil
}
