package vsgoin_node_adapter

import (
	"context"
	"s3b/vsp-blockchain/rest-api/internal/common"
	"s3b/vsp-blockchain/rest-api/internal/pb"
)

type TransactionAdapter interface {
	GenerateKeyset() (common.Keyset, error)
	GetKeysetFromWIF(privateKeyWIF string) (common.Keyset, error)
}

type TransactionAdapterImpl struct {
	appServiceClient pb.AppServiceClient
}

func NewTransactionAdapterImpl(appServiceClient pb.AppServiceClient) *TransactionAdapterImpl {

	return &TransactionAdapterImpl{
		appServiceClient: appServiceClient,
	}
}

func (t TransactionAdapterImpl) GenerateKeyset() (common.Keyset, error) {
	pbKeyset, err := t.appServiceClient.GenerateKeyset(context.Background(), nil)

	if err != nil {
		return common.Keyset{}, common.ServerError
	}

	return common.Keyset{
		PrivateKey:    [32]byte(pbKeyset.GetKeyset().PrivateKey),
		PrivateKeyWif: pbKeyset.GetKeyset().PrivateKeyWif,
		PublicKey:     [33]byte(pbKeyset.GetKeyset().PublicKey),
		VSAddress:     pbKeyset.GetKeyset().VSAddress,
	}, nil
}

func (t TransactionAdapterImpl) GetKeysetFromWIF(privateKeyWIF string) (common.Keyset, error) {
	request := pb.GetKeysetFromWIFRequest{PrivateKeyWif: privateKeyWIF}
	pbKeyset, err := t.appServiceClient.GetKeysetFromWIF(context.Background(), &request)
	if err != nil {
		return common.Keyset{}, common.ServerError
	}

	if pbKeyset.FalseInput {
		return common.Keyset{}, common.WIFInputError
	}

	return common.Keyset{
		PrivateKey:    [32]byte(pbKeyset.GetKeyset().PrivateKey),
		PrivateKeyWif: pbKeyset.GetKeyset().PrivateKeyWif,
		PublicKey:     [33]byte(pbKeyset.GetKeyset().PublicKey),
		VSAddress:     pbKeyset.GetKeyset().VSAddress,
	}, nil
}
