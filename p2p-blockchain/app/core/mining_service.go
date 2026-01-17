package core

import (
	minerApi "s3b/vsp-blockchain/p2p-blockchain/miner/api"

	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
)

// MiningService provides methods to control the mining process.
type MiningService struct {
	minerAPI minerApi.MinerAPI
}

// NewMiningService creates a new MiningService with the given miner API.
func NewMiningService(minerAPI minerApi.MinerAPI) *MiningService {
	return &MiningService{
		minerAPI: minerAPI,
	}
}

// StartMining starts the mining process with the given transactions.
func (s *MiningService) StartMining(transactions []transaction.Transaction) error {
	s.minerAPI.StartMining(transactions)
	return nil
}

// StopMining stops the ongoing mining process.
func (s *MiningService) StopMining() error {
	s.minerAPI.StopMining()
	return nil
}
