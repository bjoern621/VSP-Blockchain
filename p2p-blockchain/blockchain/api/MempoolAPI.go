package api

import (
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	"strconv"
)

type MempoolAPI struct {
	mempool *core.Mempool
}

func NewMempoolAPI(mempool *core.Mempool) *MempoolAPI {
	return &MempoolAPI{
		mempool: mempool,
	}
}

func (api *MempoolAPI) AddTransaction(transaction transaction.Transaction) bool {
	return api.mempool.AddTransaction(transaction)
}

func (api *MempoolAPI) GetTransactionValues() string {
	s := ""
	txs := api.mempool.GetTransactionsForMining()
	for _, tx := range txs {
		s += strconv.FormatUint(tx.Outputs[0].Value, 10) + "\n"
	}
	return s
}
