package vsgoin_node_adapter

type TransactionAdapter interface {
}

type TransactionAdapterImpl struct {
}

func NewTransactionAdapterImpl() *TransactionAdapterImpl {
	return &TransactionAdapterImpl{}
}
