package core

import (
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api"
)

type QueryRegistryService struct {
	queryRegistryAPI api.QueryRegistryAPI
}

func NewQueryRegistryService(queryRegistryAPI api.QueryRegistryAPI) *QueryRegistryService {
	return &QueryRegistryService{
		queryRegistryAPI: queryRegistryAPI,
	}
}

func (s *QueryRegistryService) QueryRegistry() ([]api.RegistryEntry, error) {
	return s.queryRegistryAPI.QueryRegistry()
}
